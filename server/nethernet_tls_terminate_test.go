package server

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"log/slog"
	"math/big"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// writeCertificate writes a self-signed certificate for 127.0.0.1 with serial to dir and
// returns the certificate.
func writeCertificate(t *testing.T, dir string, serial int64) *x509.Certificate {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	tmpl := &x509.Certificate{SerialNumber: big.NewInt(serial), Subject: pkix.Name{CommonName: "signaling"}, NotBefore: time.Now().Add(-time.Hour), NotAfter: time.Now().Add(time.Hour), IPAddresses: []net.IP{net.ParseIP("127.0.0.1")}, KeyUsage: x509.KeyUsageDigitalSignature, ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	keyDER, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(dir, "cert.pem"), pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}), 0600); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(dir, "key.pem"), pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER}), 0600); err != nil {
		t.Fatal(err)
	}
	cert, err := x509.ParseCertificate(der)
	if err != nil {
		t.Fatal(err)
	}
	return cert
}

// TestTerminateTLSServesBothOnOnePort checks that the signaling port answers plain HTTP and
// HTTPS with its own certificate, and serves a renewed certificate without restarting.
func TestTerminateTLSServesBothOnOnePort(t *testing.T) {
	dir := t.TempDir()
	first := writeCertificate(t, dir, 1)
	certs, err := newCertificateFiles(filepath.Join(dir, "cert.pem"), filepath.Join(dir, "key.pem"), slog.Default())
	if err != nil {
		t.Fatal(err)
	}
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	server := &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.TLS != nil {
			_, _ = w.Write([]byte("secure"))
			return
		}
		_, _ = w.Write([]byte("plain"))
	})}
	go func() { _ = server.Serve(terminateTLS(l, certs.tlsConfig(), slog.Default())) }()
	t.Cleanup(func() { _ = server.Close() })
	addr := l.Addr().String()

	if got := read(t, http.DefaultClient, "http://"+addr+"/v1/join"); got != "plain" {
		t.Fatalf("plain request answered %q", got)
	}
	client := func(cert *x509.Certificate) *http.Client {
		pool := x509.NewCertPool()
		pool.AddCert(cert)
		return &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{RootCAs: pool}, DisableKeepAlives: true}}
	}
	if got := read(t, client(first), "https://"+addr+"/v1/join"); got != "secure" {
		t.Fatalf("TLS request answered %q", got)
	}

	// A renewal: new files on disk, picked up at the next check without a restart.
	renewed := writeCertificate(t, dir, 2)
	later := time.Now().Add(time.Second)
	_ = os.Chtimes(filepath.Join(dir, "cert.pem"), later, later)
	certs.mu.Lock()
	certs.checkedAt = time.Time{}
	certs.mu.Unlock()
	if got := read(t, client(renewed), "https://"+addr+"/v1/join"); got != "secure" {
		t.Fatalf("TLS request after renewal answered %q", got)
	}
	if _, err := client(first).Get("https://" + addr + "/v1/join"); err == nil {
		t.Fatal("the replaced certificate is still served")
	}
}

// TestCertificateFilesRejectMissing checks that a missing or broken certificate fails at
// startup rather than on the first join.
func TestCertificateFilesRejectMissing(t *testing.T) {
	dir := t.TempDir()
	if _, err := newCertificateFiles(filepath.Join(dir, "cert.pem"), filepath.Join(dir, "key.pem"), slog.Default()); err == nil {
		t.Fatal("missing files accepted")
	}
	_ = os.WriteFile(filepath.Join(dir, "cert.pem"), []byte("not a certificate"), 0600)
	_ = os.WriteFile(filepath.Join(dir, "key.pem"), []byte("not a key"), 0600)
	if _, err := newCertificateFiles(filepath.Join(dir, "cert.pem"), filepath.Join(dir, "key.pem"), slog.Default()); err == nil {
		t.Fatal("broken files accepted")
	}
}
