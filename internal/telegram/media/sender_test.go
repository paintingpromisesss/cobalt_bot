package media

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"go.uber.org/zap"
	tele "gopkg.in/telebot.v4"
)

func TestResolveTelegramFileLocalModeRelativePath(t *testing.T) {
	tempDir := t.TempDir()
	t.Chdir(tempDir)

	relativePath := filepath.Join("nested", "sample.txt")
	absolutePath := filepath.Join(tempDir, relativePath)
	if err := os.MkdirAll(filepath.Dir(absolutePath), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(absolutePath, []byte("ok"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	sender := NewSender(zap.NewNop(), time.Second, time.Second, true)

	file, gotPath, err := sender.resolveTelegramFile(relativePath)
	if err != nil {
		t.Fatalf("resolveTelegramFile returned error: %v", err)
	}

	if gotPath != absolutePath {
		t.Fatalf("expected absolute path %q, got %q", absolutePath, gotPath)
	}
	if file.FileURL != "file://"+filepath.ToSlash(absolutePath) {
		t.Fatalf("expected file URL for %q, got %q", absolutePath, file.FileURL)
	}
	if file.FileLocal != "" {
		t.Fatalf("expected FileLocal to be empty in local mode, got %q", file.FileLocal)
	}
}

func TestResolveTelegramFileMultipartFallback(t *testing.T) {
	absolutePath := createTempMediaFile(t)
	sender := NewSender(zap.NewNop(), time.Second, time.Second, false)

	file, gotPath, err := sender.resolveTelegramFile(absolutePath)
	if err != nil {
		t.Fatalf("resolveTelegramFile returned error: %v", err)
	}

	if gotPath != absolutePath {
		t.Fatalf("expected absolute path %q, got %q", absolutePath, gotPath)
	}
	if file.FileLocal != absolutePath {
		t.Fatalf("expected FileLocal %q, got %q", absolutePath, file.FileLocal)
	}
	if file.FileURL != "" {
		t.Fatalf("expected FileURL to be empty in multipart mode, got %q", file.FileURL)
	}
}

func TestResolveTelegramFileValidationErrors(t *testing.T) {
	sender := NewSender(zap.NewNop(), time.Second, time.Second, true)

	if _, _, err := sender.resolveTelegramFile(""); err == nil {
		t.Fatal("expected error for empty path")
	}

	missingPath := filepath.Join(t.TempDir(), "missing.txt")
	if _, _, err := sender.resolveTelegramFile(missingPath); err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestBuildMediaUsesLocalFileURL(t *testing.T) {
	absolutePath := createTempMediaFile(t)
	sender := NewSender(zap.NewNop(), time.Second, time.Second, true)

	file, resolvedPath, err := sender.resolveTelegramFile(absolutePath)
	if err != nil {
		t.Fatalf("resolveTelegramFile returned error: %v", err)
	}

	testCases := []struct {
		name     string
		mime     string
		wantType string
	}{
		{name: "photo", mime: "image/jpeg", wantType: "photo"},
		{name: "video", mime: "video/mp4", wantType: "video"},
		{name: "audio", mime: "audio/mpeg", wantType: "audio"},
		{name: "document", mime: "application/pdf", wantType: "document"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			media, cleanup := sender.buildMedia(file, resolvedPath, "sample.bin", tc.mime)
			defer cleanup()

			switch got := media.(type) {
			case *tele.Photo:
				assertFileURL(t, got.File, tc.wantType)
			case *tele.Video:
				assertFileURL(t, got.File, tc.wantType)
			case *tele.Audio:
				assertFileURL(t, got.File, tc.wantType)
			case *tele.Document:
				assertFileURL(t, got.File, tc.wantType)
			default:
				t.Fatalf("unexpected media type %T", media)
			}
		})
	}
}

func TestBuildAlbumItemUsesLocalFileURL(t *testing.T) {
	absolutePath := createTempMediaFile(t)
	sender := NewSender(zap.NewNop(), time.Second, time.Second, true)

	item, err := sender.buildAlbumItem(absolutePath, "sample.jpg", "image/jpeg")
	if err != nil {
		t.Fatalf("buildAlbumItem returned error: %v", err)
	}

	photo, ok := item.(*tele.Photo)
	if !ok {
		t.Fatalf("expected *tele.Photo, got %T", item)
	}

	assertFileURL(t, photo.File, "album photo")
}

func createTempMediaFile(t *testing.T) string {
	t.Helper()

	tempDir := t.TempDir()
	path := filepath.Join(tempDir, "sample.bin")
	if err := os.WriteFile(path, []byte("ok"), 0o644); err != nil {
		t.Fatalf("write temp media file: %v", err)
	}
	return path
}

func assertFileURL(t *testing.T, file tele.File, label string) {
	t.Helper()

	if file.FileURL == "" {
		t.Fatalf("expected %s FileURL to be set", label)
	}
	if file.FileLocal != "" {
		t.Fatalf("expected %s FileLocal to be empty, got %q", label, file.FileLocal)
	}
}
