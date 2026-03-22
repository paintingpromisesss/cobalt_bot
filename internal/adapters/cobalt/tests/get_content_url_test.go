package tests

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	cobalt "github.com/paintingpromisesss/cobalt_bot/internal/adapters/cobalt"
)

func TestGetContentURLSuccessTunnelResponse(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected method %s, got %s", http.MethodPost, r.Method)
		}
		if got := r.Header.Get("Accept"); got != "application/json" {
			t.Fatalf("expected Accept header application/json, got %q", got)
		}
		if got := r.Header.Get("Content-Type"); got != "application/json" {
			t.Fatalf("expected Content-Type header application/json, got %q", got)
		}

		var req cobalt.MainRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}
		if req.Url != "https://youtube.com/watch?v=abc" {
			t.Fatalf("unexpected request url: %q", req.Url)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"tunnel","url":"https://cdn.example/file.mp4","filename":"file.mp4"}`))
	}))
	defer server.Close()

	client := cobalt.NewClient(server.URL, time.Second)
	got, err := client.GetContentURL(context.Background(), cobalt.MainRequest{Url: "https://youtube.com/watch?v=abc"})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if got.Status != cobalt.StatusTunnel {
		t.Fatalf("expected status %q, got %q", cobalt.StatusTunnel, got.Status)
	}
	if got.Url != "https://cdn.example/file.mp4" {
		t.Fatalf("unexpected url: %q", got.Url)
	}
	if got.Filename != "file.mp4" {
		t.Fatalf("unexpected filename: %q", got.Filename)
	}
}

func TestGetContentURLSuccessPickerResponse(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"status":"picker",
			"audio":"https://cdn.example/audio.mp3",
			"audioFilename":"audio.mp3",
			"picker":[{"type":"video","url":"https://cdn.example/video.mp4"}]
		}`))
	}))
	defer server.Close()

	client := cobalt.NewClient(server.URL, time.Second)
	got, err := client.GetContentURL(context.Background(), cobalt.MainRequest{Url: "https://example.com/post"})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if got.Status != cobalt.StatusPicker {
		t.Fatalf("expected status %q, got %q", cobalt.StatusPicker, got.Status)
	}
	if got.PickerAudio == nil || *got.PickerAudio != "https://cdn.example/audio.mp3" {
		t.Fatalf("unexpected picker audio: %#v", got.PickerAudio)
	}
	if got.AudioFilename == nil || *got.AudioFilename != "audio.mp3" {
		t.Fatalf("unexpected audio filename: %#v", got.AudioFilename)
	}
	if len(got.Picker) != 1 || got.Picker[0].Type != cobalt.PickerTypeVideo {
		t.Fatalf("unexpected picker payload: %#v", got.Picker)
	}
}

func TestGetContentURLReturnsErrorOnNon2xx(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "service unavailable", http.StatusServiceUnavailable)
	}))
	defer server.Close()

	client := cobalt.NewClient(server.URL, time.Second)
	_, err := client.GetContentURL(context.Background(), cobalt.MainRequest{Url: "https://example.com"})
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "status 503") {
		t.Fatalf("expected status code in error, got %v", err)
	}
}

func TestGetContentURLReturnsErrorOnUnsupportedStatus(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"something-unknown"}`))
	}))
	defer server.Close()

	client := cobalt.NewClient(server.URL, time.Second)
	_, err := client.GetContentURL(context.Background(), cobalt.MainRequest{Url: "https://example.com"})
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "unsupported response status") {
		t.Fatalf("expected unsupported status error, got %v", err)
	}
}
