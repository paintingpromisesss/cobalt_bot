package cobalt

import (
	"context"
	"fmt"
	"strings"
)

type CobaltClient interface {
	Process(ctx context.Context, req CobaltProcessRequest) (CobaltProcessResponse, error)
	Download(ctx context.Context, req CobaltDownloadRequest) (CobaltDownloadedFile, error)
}

type CobaltProcessRequest struct {
	URL                   string `json:"url"`
	AudioBitrate          string `json:"audioBitrate,omitempty"`
	AudioFormat           string `json:"audioFormat,omitempty"`
	DownloadMode          string `json:"downloadMode,omitempty"`
	FilenameStyle         string `json:"filenameStyle,omitempty"`
	VideoQuality          string `json:"videoQuality,omitempty"`
	DisableMetadata       bool   `json:"disableMetadata,omitempty"`
	AlwaysProxy           bool   `json:"alwaysProxy,omitempty"`
	LocalProcessing       string `json:"localProcessing,omitempty"`
	SubtitleLang          string `json:"subtitleLang,omitempty"`
	YoutubeVideoCodec     string `json:"youtubeVideoCodec,omitempty"`
	YoutubeVideoContainer string `json:"youtubeVideoContainer,omitempty"`
	YoutubeDubLang        string `json:"youtubeDubLang,omitempty"`
	ConvertGIF            bool   `json:"convertGif,omitempty"`
	AllowH265             bool   `json:"allowH265,omitempty"`
	TiktokFullAudio       bool   `json:"tiktokFullAudio,omitempty"`
	YoutubeBetterAudio    bool   `json:"youtubeBetterAudio,omitempty"`
	YoutubeHLS            bool   `json:"youtubeHLS,omitempty"`
}

func (r CobaltProcessRequest) Validate() error {
	if strings.TrimSpace(r.URL) == "" {
		return fmt.Errorf("cobalt request url is required")
	}
	return nil
}

type CobaltProcessResponse struct {
	Status   string                 `json:"status"`
	URL      string                 `json:"url,omitempty"`
	Filename string                 `json:"filename,omitempty"`
	Picker   []map[string]any       `json:"picker,omitempty"`
	Error    *CobaltResponseError   `json:"error,omitempty"`
	Raw      map[string]interface{} `json:"-"`
}

type CobaltResponseError struct {
	Code    string         `json:"code"`
	Context map[string]any `json:"context,omitempty"`
}

type CobaltAPIError struct {
	Code       string
	Context    map[string]any
	HTTPStatus int
}

func (e *CobaltAPIError) Error() string {
	if e == nil {
		return "cobalt api error"
	}
	if e.Code == "" {
		return "cobalt returned status=error"
	}
	return fmt.Sprintf("cobalt api error: %s", e.Code)
}

type CobaltHTTPError struct {
	StatusCode int
	Body       string
}

func (e *CobaltHTTPError) Error() string {
	if e == nil {
		return "cobalt http error"
	}
	if e.Body == "" {
		return fmt.Sprintf("cobalt http error: status=%d", e.StatusCode)
	}
	return fmt.Sprintf("cobalt http error: status=%d body=%s", e.StatusCode, e.Body)
}

func IsSupportedCobaltStatus(status string) bool {
	switch status {
	case "tunnel", "redirect", "local-processing", "picker", "error":
		return true
	default:
		return false
	}
}

type CobaltDownloadRequest struct {
	URL          string
	TempDir      string
	FilenameHint string
	MaxFileBytes int64
}

func (r CobaltDownloadRequest) Validate() error {
	if strings.TrimSpace(r.URL) == "" {
		return fmt.Errorf("download url is required")
	}
	if strings.TrimSpace(r.TempDir) == "" {
		return fmt.Errorf("temp dir is required")
	}
	if r.MaxFileBytes <= 0 {
		return fmt.Errorf("max file bytes must be positive, got %d", r.MaxFileBytes)
	}
	return nil
}

type CobaltDownloadedFile struct {
	Path        string
	Filename    string
	SizeBytes   int64
	ContentType string
}

type CobaltFileTooLargeError struct {
	LimitBytes int64
	ActualSize int64
}

func (e *CobaltFileTooLargeError) Error() string {
	if e == nil {
		return "cobalt downloaded file is too large"
	}
	if e.ActualSize > 0 {
		return fmt.Sprintf("cobalt downloaded file exceeds limit: actual=%d limit=%d", e.ActualSize, e.LimitBytes)
	}
	return fmt.Sprintf("cobalt downloaded file exceeds limit: limit=%d", e.LimitBytes)
}
