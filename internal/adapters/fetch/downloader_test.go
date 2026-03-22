package fetch

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDownloadSuccess(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "video/mp4")
		w.Header().Set("Content-Disposition", `attachment; filename="video.mp4"`)
		_, _ = w.Write([]byte("hello"))
	}))
	defer server.Close()

	tempDir := t.TempDir()
	d := NewDownloader(time.Second, tempDir, 1024)

	got, err := d.Download(context.Background(), server.URL+"/file", "video.mp4", nil)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if got.Filename != "video.mp4" {
		t.Fatalf("expected filename video.mp4, got %q", got.Filename)
	}
	if got.Size != 5 {
		t.Fatalf("expected size 5, got %d", got.Size)
	}
	if got.ContentType != "video/mp4" {
		t.Fatalf("unexpected content-type: %q", got.ContentType)
	}
	if got.DetectedMIME != "video/mp4" {
		t.Fatalf("unexpected detected mime: %q", got.DetectedMIME)
	}
	if filepath.Dir(got.Path) != tempDir {
		t.Fatalf("temp file is outside temp dir: %q", got.Path)
	}
	content, err := os.ReadFile(got.Path)
	if err != nil {
		t.Fatalf("failed to read downloaded file: %v", err)
	}
	if string(content) != "hello" {
		t.Fatalf("unexpected file contents: %q", string(content))
	}
}

func TestDownloadDetectMIMEFromContentWhenHeaderGeneric(t *testing.T) {
	t.Parallel()

	jpegBytes := []byte{0xFF, 0xD8, 0xFF, 0xE0, 'J', 'F', 'I', 'F', 0x00, 0x01}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")
		_, _ = w.Write(jpegBytes)
	}))
	defer server.Close()

	tempDir := t.TempDir()
	d := NewDownloader(time.Second, tempDir, 1024)

	got, err := d.Download(context.Background(), server.URL+"/img", "image.jpg", nil)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if got.ContentType != "application/octet-stream" {
		t.Fatalf("unexpected content-type: %q", got.ContentType)
	}
	if got.DetectedMIME != "image/jpeg" {
		t.Fatalf("unexpected detected mime: %q", got.DetectedMIME)
	}
}

func TestDownloadFileTooLarge(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("toolong"))
	}))
	defer server.Close()

	tempDir := t.TempDir()
	d := NewDownloader(time.Second, tempDir, 3)

	_, err := d.Download(context.Background(), server.URL, "file", nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, ErrFileTooLarge) {
		t.Fatalf("expected ErrFileTooLarge, got %v", err)
	}

	entries, readErr := os.ReadDir(tempDir)
	if readErr != nil {
		t.Fatalf("failed to read temp dir: %v", readErr)
	}
	if len(entries) != 0 {
		t.Fatalf("expected temp dir to be empty after cleanup, found %d files", len(entries))
	}
}
