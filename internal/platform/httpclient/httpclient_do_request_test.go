package httpclient

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestDoRequestSuccess(t *testing.T) {
	t.Parallel()

	type requestPayload struct {
		Name string `json:"name"`
	}
	type responsePayload struct {
		ID int `json:"id"`
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected method %s, got %s", http.MethodPost, r.Method)
		}
		if got := r.Header.Get("Accept"); got != "application/json" {
			t.Fatalf("expected Accept application/json, got %q", got)
		}
		if got := r.Header.Get("Content-Type"); got != "application/json" {
			t.Fatalf("expected Content-Type application/json, got %q", got)
		}
		if got := r.Header.Get("X-Token"); got != "token" {
			t.Fatalf("expected X-Token token, got %q", got)
		}

		var body requestPayload
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}
		if body.Name != "cobalt" {
			t.Fatalf("unexpected request body: %+v", body)
		}

		w.Header().Set("X-Request-ID", "req-1")
		_, _ = w.Write([]byte(`{"id":42}`))
	}))
	defer server.Close()

	client := New(time.Second)
	var out responsePayload
	var respHeaders http.Header
	reqHeaders := http.Header{
		"X-Token": {"token"},
	}

	err := client.DoRequest(context.Background(), Options{
		Method:          http.MethodPost,
		URL:             server.URL,
		RequestHeaders:  &reqHeaders,
		Input:           requestPayload{Name: "cobalt"},
		Output:          &out,
		ResponseHeaders: &respHeaders,
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if out.ID != 42 {
		t.Fatalf("unexpected output: %+v", out)
	}
	if respHeaders.Get("X-Request-ID") != "req-1" {
		t.Fatalf("unexpected response header %q", respHeaders.Get("X-Request-ID"))
	}
}

func TestDoRequestReturnsErrorOnNon2xx(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "bad gateway", http.StatusBadGateway)
	}))
	defer server.Close()

	client := New(time.Second)
	err := client.DoRequest(context.Background(), Options{
		Method: http.MethodGet,
		URL:    server.URL,
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "status 502") {
		t.Fatalf("expected status code in error, got %v", err)
	}
}

func TestDoRequestReturnsErrorOnDecodeFailure(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("{"))
	}))
	defer server.Close()

	client := New(time.Second)
	var out map[string]any
	err := client.DoRequest(context.Background(), Options{
		Method: http.MethodGet,
		URL:    server.URL,
		Output: &out,
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "failed to decode response") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDoRequestReturnsErrorOnMarshalFailure(t *testing.T) {
	t.Parallel()

	client := New(time.Second)
	err := client.DoRequest(context.Background(), Options{
		Method: http.MethodPost,
		URL:    "http://example.com",
		Input:  make(chan int),
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "failed to marshal input") {
		t.Fatalf("unexpected error: %v", err)
	}
}
