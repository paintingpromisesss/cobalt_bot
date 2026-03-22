package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

var ErrFileTooLarge = errors.New("file too large")

type Options struct {
	Method          string
	URL             string
	RequestHeaders  *http.Header
	Input           any
	Output          any
	ResponseHeaders *http.Header
}

type DownloadOptions struct {
	Method          string
	URL             string
	RequestHeaders  *http.Header
	ResponseHeaders *http.Header
	Output          io.Writer
	MaxBytes        int64
}

type Client struct {
	httpClient *http.Client
}

func New(timeout time.Duration) *Client {
	return &Client{
		httpClient: &http.Client{Timeout: timeout},
	}
}

func (c *Client) DoRequest(ctx context.Context, options Options) error {
	var body io.Reader
	if options.Input != nil {
		b, err := json.Marshal(options.Input)
		if err != nil {
			return fmt.Errorf("failed to marshal input: %w", err)
		}
		body = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, options.Method, options.URL, body)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	if options.RequestHeaders != nil {
		for key, values := range *options.RequestHeaders {
			for _, value := range values {
				req.Header.Add(key, value)
			}
		}
	}
	if req.Header.Get("Accept") == "" {
		req.Header.Set("Accept", "application/json")
	}
	if options.Input != nil && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		responseBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(responseBody))
	}

	if options.Output != nil {
		if err := json.NewDecoder(resp.Body).Decode(options.Output); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	if options.ResponseHeaders != nil {
		if *options.ResponseHeaders == nil {
			*options.ResponseHeaders = make(http.Header)
		}
		for key, values := range resp.Header {
			for _, value := range values {
				options.ResponseHeaders.Add(key, value)
			}
		}
	}

	return nil
}

func (c *Client) Download(ctx context.Context, options DownloadOptions) (int64, error) {
	if options.Output == nil {
		return 0, errors.New("output writer is required")
	}

	method := options.Method
	if method == "" {
		method = http.MethodGet
	}

	req, err := http.NewRequestWithContext(ctx, method, options.URL, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %w", err)
	}

	if options.RequestHeaders != nil {
		for key, values := range *options.RequestHeaders {
			for _, value := range values {
				req.Header.Add(key, value)
			}
		}
	}
	if req.Header.Get("Accept") == "" {
		req.Header.Set("Accept", "*/*")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if options.ResponseHeaders != nil {
		if *options.ResponseHeaders == nil {
			*options.ResponseHeaders = make(http.Header)
		}
		for key, values := range resp.Header {
			for _, value := range values {
				options.ResponseHeaders.Add(key, value)
			}
		}
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		responseBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return 0, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(responseBody))
	}

	reader := io.Reader(resp.Body)
	if options.MaxBytes > 0 {
		reader = io.LimitReader(resp.Body, options.MaxBytes+1)
	}

	written, err := io.Copy(options.Output, reader)
	if err != nil {
		return written, fmt.Errorf("failed to stream response body: %w", err)
	}
	if options.MaxBytes > 0 && written > options.MaxBytes {
		return written, ErrFileTooLarge
	}

	return written, nil
}
