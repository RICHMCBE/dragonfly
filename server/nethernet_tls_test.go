package server

import (
	"bufio"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// listenForward starts a plain HTTP server behind forwardTLS and returns its address.
func listenForward(t *testing.T, target string) string {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	server := &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("plain"))
	})}
	go func() { _ = server.Serve(forwardTLS(l, target, slog.Default())) }()
	t.Cleanup(func() { _ = server.Close() })
	return l.Addr().String()
}

func read(t *testing.T, client *http.Client, url string) string {
	t.Helper()
	resp, err := client.Get(url)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	return string(body)
}

func TestForwardTLSKeepsPlainHTTP(t *testing.T) {
	addr := listenForward(t, "127.0.0.1:1")
	if got := read(t, http.DefaultClient, "http://"+addr+"/v1/join"); got != "plain" {
		t.Fatalf("plain request answered %q", got)
	}
}

func TestForwardTLSReachesTarget(t *testing.T) {
	target := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("secure"))
	}))
	defer target.Close()
	addr := listenForward(t, target.Listener.Addr().String())
	if got := read(t, target.Client(), "https://"+addr+"/v1/join"); got != "secure" {
		t.Fatalf("TLS request answered %q", got)
	}
}

func TestForwardTLSSilentClientDoesNotBlockOthers(t *testing.T) {
	addr := listenForward(t, "127.0.0.1:1")
	silent, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatal(err)
	}
	defer silent.Close()
	done := make(chan string, 1)
	go func() {
		conn, err := net.Dial("tcp", addr)
		if err != nil {
			done <- err.Error()
			return
		}
		defer conn.Close()
		_, _ = conn.Write([]byte("GET / HTTP/1.0\r\n\r\n"))
		line, _ := bufio.NewReader(conn).ReadString('\n')
		done <- line
	}()
	select {
	case line := <-done:
		if line != "HTTP/1.0 200 OK\r\n" {
			t.Fatalf("got %q", line)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("silent client blocked another connection")
	}
}
