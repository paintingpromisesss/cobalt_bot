package downloader

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/paintingpromisesss/cobalt_bot/internal/httpclient"
)

var ErrFileTooLarge = errors.New("file exceeds max size")

type DownloadResult struct {
	Path        string
	Filename    string
	Size        int64
	ContentType string
}

type Downloader struct {
	httpClient   *httpclient.Client
	tempDir      string
	maxFileBytes int64
}

func NewDownloader(timeout time.Duration, tempDir string, maxFileBytes int64) *Downloader {
	return &Downloader{
		httpClient:   httpclient.New(timeout),
		tempDir:      tempDir,
		maxFileBytes: maxFileBytes,
	}
}

func (d *Downloader) Download(ctx context.Context, fileURL, filename string) (DownloadResult, error) {
	if strings.TrimSpace(fileURL) == "" {
		return DownloadResult{}, errors.New("url is required")
	}
	if strings.TrimSpace(filename) == "" {
		return DownloadResult{}, errors.New("filename is required")
	}
	if d.maxFileBytes <= 0 {
		return DownloadResult{}, errors.New("max file bytes must be positive")
	}
	if strings.TrimSpace(d.tempDir) == "" {
		return DownloadResult{}, errors.New("temp dir is required")
	}

	if err := os.MkdirAll(d.tempDir, 0o755); err != nil {
		return DownloadResult{}, fmt.Errorf("create temp dir: %w", err)
	}

	file, err := os.CreateTemp(d.tempDir, "cobalt-*")
	if err != nil {
		return DownloadResult{}, fmt.Errorf("create temp file: %w", err)
	}
	filePath := file.Name()

	cleanup := func() {
		_ = file.Close()
		_ = os.Remove(filePath)
	}

	var responseHeaders http.Header
	written, err := d.httpClient.Download(ctx, httpclient.DownloadOptions{
		Method:          http.MethodGet,
		URL:             fileURL,
		ResponseHeaders: &responseHeaders,
		Output:          file,
		MaxBytes:        d.maxFileBytes,
	})
	if err != nil {
		cleanup()
		if errors.Is(err, httpclient.ErrFileTooLarge) {
			return DownloadResult{}, ErrFileTooLarge
		}
		return DownloadResult{}, fmt.Errorf("download file: %w", err)
	}

	if err := file.Close(); err != nil {
		_ = os.Remove(filePath)
		return DownloadResult{}, fmt.Errorf("close temp file: %w", err)
	}

	return DownloadResult{
		Path:        filePath,
		Filename:    filename,
		Size:        written,
		ContentType: responseHeaders.Get("Content-Type"),
	}, nil
}
