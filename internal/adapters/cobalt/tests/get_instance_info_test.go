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

func TestGetInstanceInfoSuccess(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected method %s, got %s", http.MethodGet, r.Method)
		}
		if got := r.Header.Get("Accept"); got != "application/json" {
			t.Fatalf("expected Accept header application/json, got %q", got)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"cobalt":{
				"version":"11.5",
				"url":"https://example.com",
				"startTime":"1735689600000",
				"services":["bluesky","facebook","instagram","ok","pinterest","reddit","rutube","tiktok","twitch clips","twitter","vk","youtube"]
			},
			"git":{
				"commit":"commit_hash",
				"branch":"main",
				"remote":"imputnet/cobalt"
			}
		}`))
	}))
	defer server.Close()

	client := cobalt.NewClient(server.URL, time.Second)

	got, err := client.GetInstanceInfo(context.Background())
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if got.Cobalt.Version != "11.5" {
		t.Fatalf("unexpected cobalt.version: %q", got.Cobalt.Version)
	}
	if got.Cobalt.Url != "https://example.com" {
		t.Fatalf("unexpected cobalt.url: %q", got.Cobalt.Url)
	}
	if got.Cobalt.StartTime != "1735689600000" {
		t.Fatalf("unexpected cobalt.startTime: %q", got.Cobalt.StartTime)
	}
	if len(got.Cobalt.Services) != 12 {
		t.Fatalf("expected 12 services, got %d", len(got.Cobalt.Services))
	}
	if got.Git.Commit != "commit_hash" {
		t.Fatalf("unexpected git.commit: %q", got.Git.Commit)
	}
	if got.Git.Branch != "main" {
		t.Fatalf("unexpected git.branch: %q", got.Git.Branch)
	}
	if got.Git.Remote != "imputnet/cobalt" {
		t.Fatalf("unexpected git.remote: %q", got.Git.Remote)
	}
}

func TestGetInstanceInfoReturnsErrorOnNon2xx(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "bad gateway", http.StatusBadGateway)
	}))
	defer server.Close()

	client := cobalt.NewClient(server.URL, time.Second)

	_, err := client.GetInstanceInfo(context.Background())
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "status 502") {
		t.Fatalf("expected status code in error, got %v", err)
	}
	if !strings.Contains(err.Error(), "bad gateway") {
		t.Fatalf("expected response body in error, got %v", err)
	}
}

func TestGetInstanceInfoReturnsErrorOnInvalidJSON(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"cobalt":`))
	}))
	defer server.Close()

	client := cobalt.NewClient(server.URL, time.Second)

	_, err := client.GetInstanceInfo(context.Background())
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "failed to decode response") {
		t.Fatalf("expected decode error, got %v", err)
	}
}
