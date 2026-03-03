package httpclient

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestDoRequestStreamSuccess(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected method %s, got %s", http.MethodGet, r.Method)
		}
		if got := r.Header.Get("Accept"); got != "*/*" {
			t.Fatalf("expected Accept */*, got %q", got)
		}
		w.Header().Set("Content-Length", "5")
		_, _ = w.Write([]byte("hello"))
	}))
	defer server.Close()

	client := New(time.Second)
	var out bytes.Buffer
	var headers http.Header
	n, err := client.Download(context.Background(), DownloadOptions{
		URL:             server.URL,
		ResponseHeaders: &headers,
		Output:          &out,
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if n != 5 {
		t.Fatalf("expected 5 bytes, got %d", n)
	}
	if out.String() != "hello" {
		t.Fatalf("unexpected body %q", out.String())
	}
	if headers.Get("Content-Length") != "5" {
		t.Fatalf("unexpected content-length header %q", headers.Get("Content-Length"))
	}
}

func TestDoRequestStreamMaxBytesExceeded(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("toolong"))
	}))
	defer server.Close()

	client := New(time.Second)
	var out bytes.Buffer
	_, err := client.Download(context.Background(), DownloadOptions{
		URL:      server.URL,
		Output:   &out,
		MaxBytes: 3,
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "exceeds max bytes limit") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDoRequestStreamReturnsErrorOnNon2xx(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "bad gateway", http.StatusBadGateway)
	}))
	defer server.Close()

	client := New(time.Second)
	var out bytes.Buffer
	_, err := client.Download(context.Background(), DownloadOptions{
		URL:    server.URL,
		Output: &out,
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "status 502") {
		t.Fatalf("expected status code in error, got %v", err)
	}
}
