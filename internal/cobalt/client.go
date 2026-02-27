package cobalt

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	neturl "net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

const maxErrorBodyBytes = 4096

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient(baseURL string, timeout time.Duration) (*Client, error) {
	baseURL = strings.TrimSpace(baseURL)
	if baseURL == "" {
		return nil, fmt.Errorf("cobalt base url is empty")
	}
	if timeout <= 0 {
		return nil, fmt.Errorf("cobalt timeout must be positive, got %s", timeout)
	}
	if !strings.HasSuffix(baseURL, "/") {
		baseURL += "/"
	}

	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}, nil
}

func (c *Client) Process(ctx context.Context, req CobaltProcessRequest) (CobaltProcessResponse, error) {
	if err := req.Validate(); err != nil {
		return CobaltProcessResponse{}, err
	}

	payload, err := json.Marshal(req)
	if err != nil {
		return CobaltProcessResponse{}, fmt.Errorf("marshal cobalt request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL, bytes.NewReader(payload))
	if err != nil {
		return CobaltProcessResponse{}, fmt.Errorf("create cobalt request: %w", err)
	}
	httpReq.Header.Set("Accept", "application/json")
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return CobaltProcessResponse{}, fmt.Errorf("send cobalt request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return CobaltProcessResponse{}, fmt.Errorf("read cobalt response: %w", err)
	}

	var out CobaltProcessResponse
	if len(bytes.TrimSpace(body)) > 0 {
		if err := json.Unmarshal(body, &out); err != nil {
			if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
				return CobaltProcessResponse{}, &CobaltHTTPError{
					StatusCode: resp.StatusCode,
					Body:       truncateBody(body, maxErrorBodyBytes),
				}
			}
			return CobaltProcessResponse{}, fmt.Errorf("decode cobalt response: %w", err)
		}
	}

	if out.Status == "error" {
		apiErr := &CobaltAPIError{HTTPStatus: resp.StatusCode}
		if out.Error != nil {
			apiErr.Code = out.Error.Code
			apiErr.Context = out.Error.Context
		}
		return out, apiErr
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return out, &CobaltHTTPError{
			StatusCode: resp.StatusCode,
			Body:       truncateBody(body, maxErrorBodyBytes),
		}
	}

	if !IsSupportedCobaltStatus(out.Status) {
		return out, fmt.Errorf("unsupported cobalt response status: %q", out.Status)
	}

	return out, nil
}

func (c *Client) Download(ctx context.Context, req CobaltDownloadRequest) (CobaltDownloadedFile, error) {
	if err := req.Validate(); err != nil {
		return CobaltDownloadedFile{}, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, req.URL, nil)
	if err != nil {
		return CobaltDownloadedFile{}, fmt.Errorf("create cobalt download request: %w", err)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return CobaltDownloadedFile{}, fmt.Errorf("send cobalt download request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, maxErrorBodyBytes+1))
		return CobaltDownloadedFile{}, &CobaltHTTPError{
			StatusCode: resp.StatusCode,
			Body:       truncateBody(body, maxErrorBodyBytes),
		}
	}

	if resp.ContentLength > 0 && resp.ContentLength > req.MaxFileBytes {
		return CobaltDownloadedFile{}, &CobaltFileTooLargeError{
			LimitBytes: req.MaxFileBytes,
			ActualSize: resp.ContentLength,
		}
	}

	if err := os.MkdirAll(req.TempDir, 0o755); err != nil {
		return CobaltDownloadedFile{}, fmt.Errorf("create temp dir: %w", err)
	}

	filename := resolveFilename(req.URL, req.FilenameHint, resp.Header.Get("Content-Disposition"))
	pattern := "cobalt-*"
	if ext := filepath.Ext(filename); ext != "" {
		pattern += ext
	}

	tmpFile, err := os.CreateTemp(req.TempDir, pattern)
	if err != nil {
		return CobaltDownloadedFile{}, fmt.Errorf("create temp file: %w", err)
	}

	var (
		totalBytes int64
		writeErr   error
	)

	buf := make([]byte, 32*1024)
	for {
		n, readErr := resp.Body.Read(buf)
		if n > 0 {
			totalBytes += int64(n)
			if totalBytes > req.MaxFileBytes {
				writeErr = &CobaltFileTooLargeError{
					LimitBytes: req.MaxFileBytes,
					ActualSize: totalBytes,
				}
				break
			}

			if _, err := tmpFile.Write(buf[:n]); err != nil {
				writeErr = fmt.Errorf("write temp file: %w", err)
				break
			}
		}

		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			writeErr = fmt.Errorf("read download stream: %w", readErr)
			break
		}
	}

	if closeErr := tmpFile.Close(); closeErr != nil && writeErr == nil {
		writeErr = fmt.Errorf("close temp file: %w", closeErr)
	}

	if writeErr != nil {
		_ = os.Remove(tmpFile.Name())
		return CobaltDownloadedFile{}, writeErr
	}

	return CobaltDownloadedFile{
		Path:        tmpFile.Name(),
		Filename:    filename,
		SizeBytes:   totalBytes,
		ContentType: resp.Header.Get("Content-Type"),
	}, nil
}

func truncateBody(body []byte, max int) string {
	if len(body) <= max {
		return string(body)
	}
	return string(body[:max]) + "...(truncated)"
}

func resolveFilename(rawURL, hint, contentDisposition string) string {
	if sanitized := sanitizeFilename(hint); sanitized != "" {
		return sanitized
	}

	if contentDisposition != "" {
		if _, params, err := mime.ParseMediaType(contentDisposition); err == nil {
			if name := sanitizeFilename(params["filename"]); name != "" {
				return name
			}
		}
	}

	if parsed, err := neturl.Parse(rawURL); err == nil {
		if base := sanitizeFilename(path.Base(parsed.Path)); base != "" && base != "." && base != "/" {
			return base
		}
	}

	return "download.bin"
}

func sanitizeFilename(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return ""
	}
	name = filepath.Base(name)
	if name == "." || name == "/" {
		return ""
	}

	clean := strings.Map(func(r rune) rune {
		if r < 32 {
			return -1
		}
		switch r {
		case '/', '\\', ':', '*', '?', '"', '<', '>', '|':
			return '_'
		default:
			return r
		}
	}, name)

	clean = strings.TrimSpace(clean)
	if clean == "" || clean == "." {
		return ""
	}
	return clean
}
