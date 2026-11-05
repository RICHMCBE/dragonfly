package server

import (
	"bufio"
	"io"
	"log/slog"
	"net"
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
type tlsForwardListener struct {
	net.Listener
	target   string
	log      *slog.Logger
	accepted chan net.Conn
	closed   chan struct{}
	once     sync.Once
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
	buffered := &bufferedConn{Conn: conn, reader: reader}
	if first[0] == tlsHandshakeRecord {
		f.forward(buffered)
		return
	}
	select {
	case f.accepted <- buffered:
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
