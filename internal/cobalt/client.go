package cobalt

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/paintingpromisesss/cobalt_bot/internal/httpclient"
)

type CobaltClient struct {
	baseURL    string
	httpClient *httpclient.Client
}

func NewCobaltClient(baseURL string, timeout time.Duration) *CobaltClient {
	return &CobaltClient{
		baseURL:    baseURL,
		httpClient: httpclient.New(timeout),
	}
}

func (c *CobaltClient) GetContentURL(ctx context.Context, request MainRequest) (MainResponse, error) {
	var output json.RawMessage
	requestHeaders := http.Header{}
	requestHeaders.Set("Accept", "application/json")
	requestHeaders.Set("Content-Type", "application/json")

	if err := c.httpClient.DoRequest(ctx, httpclient.Options{
		Method:         http.MethodPost,
		URL:            c.baseURL,
		RequestHeaders: &requestHeaders,
		Input:          request,
		Output:         &output,
	}); err != nil {
		return MainResponse{}, fmt.Errorf("cobalt request failed: %w", err)
	}

	return ParseMainResponse(output)
}

func (c *CobaltClient) GetEstimatedFileSizeByURL(ctx context.Context, url string) (int64, error) {
	var responseHeaders http.Header
	requestHeaders := http.Header{}
	requestHeaders.Set("Accept", "*/*")

	if err := c.httpClient.DoRequest(ctx, httpclient.Options{
		Method:          http.MethodGet,
		URL:             url,
		RequestHeaders:  &requestHeaders,
		ResponseHeaders: &responseHeaders,
	}); err != nil {
		return 0, fmt.Errorf("cobalt request failed: %w", err)
	}
	contentLength := responseHeaders.Get("Content-Length")
	if contentLength == "" {
		return 0, fmt.Errorf("Content-Length header not found")
	}
	fileSize, err := strconv.ParseInt(contentLength, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse Content-Length: %w", err)
	}
	return fileSize, nil
}
