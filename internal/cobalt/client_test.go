package cobalt

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

func TestNewClientValidation(t *testing.T) {
	if _, err := NewClient("", 5*time.Second); err == nil {
		t.Fatalf("expected error for empty base url")
	}

	if _, err := NewClient("http://localhost:9000", 0); err == nil {
		t.Fatalf("expected error for non-positive timeout")
	}
}

func TestClientProcessSendsHeadersAndParsesTunnelResponse(t *testing.T) {
	var gotAccept string
	var gotContentType string
	var gotReq CobaltProcessRequest

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAccept = r.Header.Get("Accept")
		gotContentType = r.Header.Get("Content-Type")

		if err := json.NewDecoder(r.Body).Decode(&gotReq); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		_ = json.NewEncoder(w).Encode(CobaltProcessResponse{
			Status:   "tunnel",
			URL:      "http://cobalt:9000/tunnel?id=abc",
			Filename: "video.mp4",
		})
	}))
	defer srv.Close()

	client, err := NewClient(srv.URL, 5*time.Second)
	if err != nil {
		t.Fatalf("init client: %v", err)
	}

	resp, err := client.Process(context.Background(), CobaltProcessRequest{
		URL:          "https://example.com/video",
		VideoQuality: "1080",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotAccept != "application/json" {
		t.Fatalf("unexpected Accept header: %q", gotAccept)
	}
	if gotContentType != "application/json" {
		t.Fatalf("unexpected Content-Type header: %q", gotContentType)
	}
	if gotReq.URL != "https://example.com/video" {
		t.Fatalf("unexpected request url: %q", gotReq.URL)
	}
	if gotReq.VideoQuality != "1080" {
		t.Fatalf("unexpected request video quality: %q", gotReq.VideoQuality)
	}
	if resp.Status != "tunnel" || resp.URL == "" || resp.Filename != "video.mp4" {
		t.Fatalf("unexpected response: %+v", resp)
	}
}

func TestClientProcessReturnsAPIErrorForStatusError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(CobaltProcessResponse{
			Status: "error",
			Error: &CobaltResponseError{
				Code: "error.api.link.invalid",
				Context: map[string]any{
					"url": "bad",
				},
			},
		})
	}))
	defer srv.Close()

	client, _ := NewClient(srv.URL, 5*time.Second)
	resp, err := client.Process(context.Background(), CobaltProcessRequest{
		URL: "https://bad.example",
	})
	if err == nil {
		t.Fatalf("expected cobalt api error")
	}

	var apiErr *CobaltAPIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *service.CobaltAPIError, got %T", err)
	}
	if apiErr.Code != "error.api.link.invalid" {
		t.Fatalf("unexpected api error code: %q", apiErr.Code)
	}
	if resp.Status != "error" {
		t.Fatalf("expected response status=error, got %q", resp.Status)
	}
}

func TestClientProcessReturnsHTTPErrorForNon2xx(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte("rate limited"))
	}))
	defer srv.Close()

	client, _ := NewClient(srv.URL, 5*time.Second)
	_, err := client.Process(context.Background(), CobaltProcessRequest{
		URL: "https://example.com/video",
	})
	if err == nil {
		t.Fatalf("expected http error")
	}

	var httpErr *CobaltHTTPError
	if !errors.As(err, &httpErr) {
		t.Fatalf("expected *service.CobaltHTTPError, got %T", err)
	}
	if httpErr.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("unexpected status code: %d", httpErr.StatusCode)
	}
}

func TestClientProcessReturnsErrorForUnsupportedStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(CobaltProcessResponse{
			Status: "mystery",
		})
	}))
	defer srv.Close()

	client, _ := NewClient(srv.URL, 5*time.Second)
	_, err := client.Process(context.Background(), CobaltProcessRequest{
		URL: "https://example.com/video",
	})
	if err == nil {
		t.Fatalf("expected unsupported status error")
	}
}

func TestClientProcessValidatesRequest(t *testing.T) {
	client, _ := NewClient("http://localhost:9000", 5*time.Second)

	_, err := client.Process(context.Background(), CobaltProcessRequest{})
	if err == nil {
		t.Fatalf("expected validation error for empty url")
	}
}

func TestClientDownloadSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "video/mp4")
		_, _ = w.Write([]byte("hello"))
	}))
	defer srv.Close()

	client, _ := NewClient(srv.URL, 5*time.Second)
	resp, err := client.Download(context.Background(), CobaltDownloadRequest{
		URL:          srv.URL + "/tunnel?id=1",
		TempDir:      t.TempDir(),
		FilenameHint: "video.mp4",
		MaxFileBytes: 1024,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(resp.Path)
	if err != nil {
		t.Fatalf("read downloaded file: %v", err)
	}
	if string(data) != "hello" {
		t.Fatalf("unexpected downloaded content: %q", string(data))
	}
	if resp.Filename != "video.mp4" {
		t.Fatalf("unexpected filename: %q", resp.Filename)
	}
	if resp.SizeBytes != 5 {
		t.Fatalf("unexpected size: %d", resp.SizeBytes)
	}
}

func TestClientDownloadUsesContentDispositionFilename(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Disposition", `attachment; filename="clip.mp4"`)
		_, _ = w.Write([]byte("ok"))
	}))
	defer srv.Close()

	client, _ := NewClient(srv.URL, 5*time.Second)
	resp, err := client.Download(context.Background(), CobaltDownloadRequest{
		URL:          srv.URL + "/tunnel?id=2",
		TempDir:      t.TempDir(),
		MaxFileBytes: 1024,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Filename != "clip.mp4" {
		t.Fatalf("expected filename from content-disposition, got %q", resp.Filename)
	}
}

func TestClientDownloadRejectsByContentLength(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "100")
		_, _ = w.Write([]byte("body"))
	}))
	defer srv.Close()

	client, _ := NewClient(srv.URL, 5*time.Second)
	_, err := client.Download(context.Background(), CobaltDownloadRequest{
		URL:          srv.URL + "/tunnel?id=3",
		TempDir:      t.TempDir(),
		MaxFileBytes: 10,
	})
	if err == nil {
		t.Fatalf("expected too large error")
	}

	var tooLargeErr *CobaltFileTooLargeError
	if !errors.As(err, &tooLargeErr) {
		t.Fatalf("expected *service.CobaltFileTooLargeError, got %T", err)
	}
	if tooLargeErr.LimitBytes != 10 {
		t.Fatalf("unexpected limit in error: %d", tooLargeErr.LimitBytes)
	}
}

func TestClientDownloadRejectsByStreamSize(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(strings.Repeat("a", 20)))
	}))
	defer srv.Close()

	tmpDir := t.TempDir()
	client, _ := NewClient(srv.URL, 5*time.Second)
	_, err := client.Download(context.Background(), CobaltDownloadRequest{
		URL:          srv.URL + "/tunnel?id=4",
		TempDir:      tmpDir,
		MaxFileBytes: 5,
	})
	if err == nil {
		t.Fatalf("expected too large error")
	}

	var tooLargeErr *CobaltFileTooLargeError
	if !errors.As(err, &tooLargeErr) {
		t.Fatalf("expected *service.CobaltFileTooLargeError, got %T", err)
	}

	entries, readErr := os.ReadDir(tmpDir)
	if readErr != nil {
		t.Fatalf("read tmp dir: %v", readErr)
	}
	if len(entries) != 0 {
		t.Fatalf("expected temp file cleanup on error, got %d entries", len(entries))
	}
}

func TestClientDownloadReturnsHTTPErrorForNon2xx(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte("denied"))
	}))
	defer srv.Close()

	client, _ := NewClient(srv.URL, 5*time.Second)
	_, err := client.Download(context.Background(), CobaltDownloadRequest{
		URL:          srv.URL + "/tunnel?id=5",
		TempDir:      t.TempDir(),
		MaxFileBytes: 1024,
	})
	if err == nil {
		t.Fatalf("expected http error")
	}

	var httpErr *CobaltHTTPError
	if !errors.As(err, &httpErr) {
		t.Fatalf("expected *service.CobaltHTTPError, got %T", err)
	}
	if httpErr.StatusCode != http.StatusForbidden {
		t.Fatalf("unexpected status code: %d", httpErr.StatusCode)
	}
}

func TestClientDownloadValidatesRequest(t *testing.T) {
	client, _ := NewClient("http://localhost:9000", 5*time.Second)
	_, err := client.Download(context.Background(), CobaltDownloadRequest{})
	if err == nil {
		t.Fatalf("expected validation error")
	}
}
