package cobalt

import (
	"context"
	"fmt"
	"net/http"

	"github.com/paintingpromisesss/cobalt_bot/internal/platform/httpclient"
)

type CobaltObject struct {
	Version          string   `json:"version"`
	Url              string   `json:"url"`
	StartTime        string   `json:"startTime"`
	TurnstileSitekey string   `json:"turnstileSitekey,omitempty"`
	Services         []string `json:"services"`
}

type GitObject struct {
	Commit string `json:"commit"`
	Branch string `json:"branch"`
	Remote string `json:"remote"`
}

type InstanceResponse struct {
	Cobalt CobaltObject `json:"cobalt"`
	Git    GitObject    `json:"git"`
}

func (c *Client) GetInstanceInfo(ctx context.Context) (InstanceResponse, error) {
	var instanceInfo InstanceResponse
	headers := http.Header{}
	headers.Set("Accept", "application/json")

	err := c.httpClient.DoRequest(ctx, httpclient.Options{
		Method:         http.MethodGet,
		URL:            c.baseURL,
		RequestHeaders: &headers,
		Output:         &instanceInfo,
	})
	if err != nil {
		return InstanceResponse{}, fmt.Errorf("cobalt request failed: %w", err)
	}
	return instanceInfo, err
}
