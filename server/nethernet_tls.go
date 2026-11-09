package server

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"
	"sync"
	"time"
)

const (
	// tlsHandshakeRecord is the first byte of every TLS connection.
	tlsHandshakeRecord = 0x16
	// tlsSniffTimeout limits how long a connection may stay silent before it is dropped.
	tlsSniffTimeout = 5 * time.Second
	// tlsForwardLifetime limits a forwarded connection. Signaling is a few short requests.
	tlsForwardLifetime = time.Minute
)

// tlsForwardListener accepts plaintext connections like the listener it wraps, and hands
// those that open with a TLS handshake to target untouched. It lets the signaling port
// answer HTTPS through a server that holds a certificate, while plain HTTP keeps working
// for players joining by IP address.
//
// With a TLS configuration instead of a target, it terminates TLS itself and hands out the
// decrypted connections, so no other server or port is needed for HTTPS.
type tlsForwardListener struct {
	net.Listener
	target   string
	tls      *tls.Config
	log      *slog.Logger
	accepted chan net.Conn
	closed   chan struct{}
	once     sync.Once
}

// terminateTLS wraps l so that connections opening with a TLS handshake are decrypted with
// config and come out of Accept like the plaintext ones.
func terminateTLS(l net.Listener, config *tls.Config, log *slog.Logger) net.Listener {
	f := &tlsForwardListener{Listener: l, tls: config, log: log, accepted: make(chan net.Conn), closed: make(chan struct{})}
	go f.acceptLoop()
	return f
}

// forwardTLS wraps l so that TLS connections go to target and the others come out of Accept.
func forwardTLS(l net.Listener, target string, log *slog.Logger) net.Listener {
	f := &tlsForwardListener{Listener: l, target: target, log: log, accepted: make(chan net.Conn), closed: make(chan struct{})}
	go f.acceptLoop()
	return f
}

func (f *tlsForwardListener) acceptLoop() {
	for {
		conn, err := f.Listener.Accept()
		if err != nil {
			_ = f.Close()
			return
		}
		go f.sniff(conn)
	}
}

// sniff reads the first byte of conn on its own goroutine, so that a silent client cannot
// hold up the others.
func (f *tlsForwardListener) sniff(conn net.Conn) {
	_ = conn.SetReadDeadline(time.Now().Add(tlsSniffTimeout))
	reader := bufio.NewReader(conn)
	first, err := reader.Peek(1)
	if err != nil {
		_ = conn.Close()
		return
	}
	_ = conn.SetReadDeadline(time.Time{})
	var out net.Conn = &bufferedConn{Conn: conn, reader: reader}
	if first[0] == tlsHandshakeRecord {
		if f.tls == nil {
			f.forward(out)
			return
		}
		secure := tls.Server(out, f.tls)
		_ = secure.SetDeadline(time.Now().Add(tlsSniffTimeout))
		if err := secure.Handshake(); err != nil {
			f.log.Debug("NetherNet TLS handshake failed.", "remote", conn.RemoteAddr(), "error", err)
			_ = conn.Close()
			return
		}
		_ = secure.SetDeadline(time.Time{})
		out = secure
	}
	select {
	case f.accepted <- out:
	case <-f.closed:
		_ = conn.Close()
	}
}

func (f *tlsForwardListener) forward(conn net.Conn) {
	defer conn.Close()
	remote, err := net.DialTimeout("tcp", f.target, tlsSniffTimeout)
	if err != nil {
		f.log.Warn("Could not reach the NetherNet TLS target.", "target", f.target, "error", err)
		return
	}
	defer remote.Close()
	deadline := time.Now().Add(tlsForwardLifetime)
	_ = conn.SetDeadline(deadline)
	_ = remote.SetDeadline(deadline)

	done := make(chan struct{}, 2)
	pipe := func(dst, src net.Conn) {
		_, _ = io.Copy(dst, src)
		if tcp, ok := dst.(*net.TCPConn); ok {
			_ = tcp.CloseWrite()
		}
		done <- struct{}{}
	}
	go pipe(remote, conn)
	go pipe(tcpOf(conn), remote)
	<-done
	<-done
}

// tcpOf returns the TCP connection under conn, so that half-closing reaches the client.
func tcpOf(conn net.Conn) net.Conn {
	if b, ok := conn.(*bufferedConn); ok {
		return b.Conn
	}
	return conn
}

// Accept returns the next plaintext connection.
func (f *tlsForwardListener) Accept() (net.Conn, error) {
	select {
	case conn := <-f.accepted:
		return conn, nil
	case <-f.closed:
		return nil, net.ErrClosed
	}
}

// Close stops accepting connections and closes the wrapped listener.
func (f *tlsForwardListener) Close() error {
	var err error
	f.once.Do(func() {
		close(f.closed)
		err = f.Listener.Close()
	})
	return err
}

// bufferedConn is a net.Conn that first gives back what sniffing read from it.
type bufferedConn struct {
	net.Conn
	reader *bufio.Reader
}

// Read reads from the buffer, then the connection.
func (b *bufferedConn) Read(p []byte) (int, error) {
	return b.reader.Read(p)
}

// certificateCheckInterval is how often the certificate files are checked for a renewal.
const certificateCheckInterval = time.Minute

// certificateFiles serves a certificate from PEM files and loads it again once the files
// change, so a renewed certificate is picked up without restarting or dropping players.
type certificateFiles struct {
	certFile, keyFile string
	log               *slog.Logger

	mu        sync.Mutex
	cert      *tls.Certificate
	certMod   time.Time
	keyMod    time.Time
	checkedAt time.Time
}

// newCertificateFiles loads the certificate in certFile and keyFile, failing if it cannot.
func newCertificateFiles(certFile, keyFile string, log *slog.Logger) (*certificateFiles, error) {
	c := &certificateFiles{certFile: certFile, keyFile: keyFile, log: log}
	if err := c.load(); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *certificateFiles) load() error {
	certInfo, err := os.Stat(c.certFile)
	if err != nil {
		return fmt.Errorf("NetherNet TLS certificate: %w", err)
	}
	keyInfo, err := os.Stat(c.keyFile)
	if err != nil {
		return fmt.Errorf("NetherNet TLS key: %w", err)
	}
	cert, err := tls.LoadX509KeyPair(c.certFile, c.keyFile)
	if err != nil {
		return fmt.Errorf("NetherNet TLS key pair: %w", err)
	}
	c.cert, c.certMod, c.keyMod = &cert, certInfo.ModTime(), keyInfo.ModTime()
	return nil
}

// getCertificate returns the current certificate, reloading it at most every
// certificateCheckInterval when either file changed. A failed reload keeps the old one.
func (c *certificateFiles) getCertificate(*tls.ClientHelloInfo) (*tls.Certificate, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if now := time.Now(); now.Sub(c.checkedAt) >= certificateCheckInterval {
		c.checkedAt = now
		certInfo, certErr := os.Stat(c.certFile)
		keyInfo, keyErr := os.Stat(c.keyFile)
		if certErr == nil && keyErr == nil && (!certInfo.ModTime().Equal(c.certMod) || !keyInfo.ModTime().Equal(c.keyMod)) {
			if err := c.load(); err != nil {
				c.log.Warn("Could not reload the NetherNet TLS certificate; keeping the previous one.", "error", err)
			} else {
				c.log.Info("Reloaded the NetherNet TLS certificate.", "file", c.certFile)
			}
		}
	}
	return c.cert, nil
}

// tlsConfig returns a TLS configuration serving c for HTTP/1.1 signaling.
func (c *certificateFiles) tlsConfig() *tls.Config {
	return &tls.Config{GetCertificate: c.getCertificate, NextProtos: []string{"http/1.1"}, MinVersion: tls.VersionTLS12}
}
