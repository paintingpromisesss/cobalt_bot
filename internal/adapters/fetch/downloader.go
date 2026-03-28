package fetch

import (
	"context"
	"errors"
	"fmt"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/paintingpromisesss/cobalt_bot/internal/platform/httpclient"
)

var ErrFileTooLarge = errors.New("file exceeds max size")
var ErrEmptyFile = errors.New("downloaded file is empty")

type DownloadResult struct {
	Path         string
	Filename     string
	Size         int64
	ContentType  string
	DetectedMIME string
}

type Downloader struct {
	httpClient   *httpclient.Client
	tempDir      string
	maxFileBytes int64
}

type MultiDownloadFiles struct {
	URL      string
	Filename string
}

func NewDownloader(timeout time.Duration, tempDir string, maxFileBytes int64) *Downloader {
	return &Downloader{
		httpClient:   httpclient.New(timeout),
		tempDir:      tempDir,
		maxFileBytes: maxFileBytes,
	}
}

func (d *Downloader) Download(ctx context.Context, fileURL, filename string, requestHeaders *http.Header) (DownloadResult, error) {
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

	file, err := os.CreateTemp(d.tempDir, "temp-")
	if err != nil {
		return DownloadResult{}, fmt.Errorf("create temp file: %w", err)
	}
	filePath := file.Name()

	cleanup := func() {
		_ = closeFile(file)
		_ = os.Remove(filePath)
	}
	var responseHeaders http.Header
	written, err := d.httpClient.Download(ctx, httpclient.DownloadOptions{
		Method:          http.MethodGet,
		URL:             fileURL,
		ResponseHeaders: &responseHeaders,
		RequestHeaders:  requestHeaders,
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
	if written == 0 {
		cleanup()
		return DownloadResult{}, ErrEmptyFile
	}

	if err := file.Close(); err != nil {
		_ = os.Remove(filePath)
		return DownloadResult{}, fmt.Errorf("close temp file: %w", err)
	}
	if err := os.Chmod(filePath, 0o644); err != nil {
		_ = os.Remove(filePath)
		return DownloadResult{}, fmt.Errorf("set temp file permissions: %w", err)
	}

	return DownloadResult{
		Path:         filePath,
		Filename:     filename,
		Size:         written,
		ContentType:  responseHeaders.Get("Content-Type"),
		DetectedMIME: detectMIME(filePath, responseHeaders.Get("Content-Type"), filename),
	}, nil
}

func (d *Downloader) MultiDownload(ctx context.Context, files []MultiDownloadFiles, requestHeaders *http.Header) ([]DownloadResult, error) {
	results := make([]DownloadResult, 0, len(files))
	for _, file := range files {
		result, err := d.Download(ctx, file.URL, file.Filename, requestHeaders)
		if err != nil {
			return results, fmt.Errorf("download file %s: %w", file.Filename, err)
		}
		results = append(results, result)
	}
	return results, nil
}

func detectMIME(filePath, contentTypeHeader, filename string) string {
	headerMIME := normalizeMIME(contentTypeHeader)
	if headerMIME != "" && headerMIME != "application/octet-stream" {
		return headerMIME
	}

	var sniffMIME string
	if file, err := os.Open(filePath); err == nil {
		defer func() {
			_ = closeFile(file)
		}()
		buf := make([]byte, 512)
		if n, readErr := file.Read(buf); readErr == nil || n > 0 {
			sniffMIME = normalizeMIME(http.DetectContentType(buf[:n]))
		}
	}
	if sniffMIME != "" && sniffMIME != "application/octet-stream" {
		return sniffMIME
	}

	if headerMIME != "" {
		return headerMIME
	}

	extMIME := normalizeMIME(mime.TypeByExtension(strings.ToLower(filepath.Ext(filename))))
	if extMIME != "" {
		return extMIME
	}

	if sniffMIME != "" {
		return sniffMIME
	}

	return "application/octet-stream"
}

func normalizeMIME(value string) string {
	if value == "" {
		return ""
	}

	parts := strings.Split(value, ";")
	return strings.TrimSpace(parts[0])
}

func closeFile(file *os.File) error {
	if file == nil {
		return nil
	}
	return file.Close()
}
