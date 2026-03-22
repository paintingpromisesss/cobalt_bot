package tests

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	cobalt "github.com/paintingpromisesss/cobalt_bot/internal/adapters/cobalt"
)

func TestGetEstimatedFileSizeByURLSuccess(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected method %s, got %s", http.MethodGet, r.Method)
		}
		if got := r.Header.Get("Accept"); got != "*/*" {
			t.Fatalf("expected Accept header */*, got %q", got)
		}
		w.Header().Set("Content-Length", "12345")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := cobalt.NewClient("http://unused.local", time.Second)
	got, err := client.GetEstimatedFileSizeByURL(context.Background(), server.URL)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if got != 12345 {
		t.Fatalf("expected file size 12345, got %d", got)
	}
}

func TestGetEstimatedFileSizeByURLReturnsErrorOnNon2xx(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "bad gateway", http.StatusBadGateway)
	}))
	defer server.Close()

	client := cobalt.NewClient("http://unused.local", time.Second)
	_, err := client.GetEstimatedFileSizeByURL(context.Background(), server.URL)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "status 502") {
		t.Fatalf("expected status code in error, got %v", err)
	}
}
