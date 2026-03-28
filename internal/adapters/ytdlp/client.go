package ytdlp

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"mime"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/paintingpromisesss/cobalt_bot/internal/adapters/fetch"
	ffprobe "github.com/paintingpromisesss/cobalt_bot/internal/adapters/ffprobe"
	"github.com/paintingpromisesss/cobalt_bot/internal/domain/media"
)

var ErrMediaDurationTooLong = errors.New("media duration exceeds limit")

type Client struct {
	tempDir                string
	MaxDurationSecs        int
	MaxFileBytes           int64
	CurrentlyLiveAvailable bool
	PlaylistAvailable      bool
	ClientType             *YtDLPClient
	JSRuntimeSpec          string
}

func NewClient(tempDir string, maxDurationSecs int, maxFileBytes int64, currentlyLiveAvailable bool, playlistAvailable bool, useJSRuntime bool) *Client {
	normalizedTempDir := normalizeTempDir(tempDir)

	return &Client{
		tempDir:                normalizedTempDir,
		MaxDurationSecs:        maxDurationSecs,
		MaxFileBytes:           maxFileBytes,
		CurrentlyLiveAvailable: currentlyLiveAvailable,
		PlaylistAvailable:      playlistAvailable,
		JSRuntimeSpec:          detectJSRuntimeSpec(useJSRuntime),
	}
}

func normalizeTempDir(tempDir string) string {
	trimmed := strings.TrimSpace(tempDir)
	if trimmed == "" {
		return ""
	}

	absolutePath, err := filepath.Abs(trimmed)
	if err != nil {
		return trimmed
	}

	return absolutePath
}

func (c *Client) prepareRuntimeDirectories() error {
	if strings.TrimSpace(c.tempDir) == "" {
		return nil
	}

	directories := []string{
		c.tempDir,
		filepath.Join(c.tempDir, ".home"),
		filepath.Join(c.tempDir, ".cache"),
		filepath.Join(c.tempDir, ".parts"),
	}

	for _, directory := range directories {
		if err := os.MkdirAll(directory, 0o755); err != nil {
			return fmt.Errorf("create yt-dlp runtime directory %s: %w", directory, err)
		}
	}

	return nil
}

func (c *Client) GetMetadata(ctx context.Context, url string) (*Metadata, error) {
	args := c.buildGetMetadataArgs(url, c.ClientType)
	cmd := exec.CommandContext(ctx, "yt-dlp", args...)

	if err := c.prepareRuntimeDirectories(); err != nil {
		return nil, err
	}
	cmd.Env = c.defaultEnvironment()
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var metadata Metadata
	if err := json.Unmarshal(output, &metadata); err != nil {
		return nil, err
	}
	if err := validateMediaDurationSeconds(metadata.Duration, c.MaxDurationSecs); err != nil {
		return nil, err
	}

	return &metadata, nil
}

func (c *Client) Download(ctx context.Context, url, formatID string, selectedFormat *media.DownloadFormat) (*fetch.DownloadResult, error) {
	args := c.buildDownloadArgs(url, formatID, selectedFormat)
	cmd := exec.CommandContext(ctx, "yt-dlp", args...)

	if err := c.prepareRuntimeDirectories(); err != nil {
		return nil, err
	}
	cmd.Env = c.defaultEnvironment()
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	output, err := cmd.Output()
	if err != nil {
		errText := strings.TrimSpace(stderr.String())
		if errText == "" {
			return nil, fmt.Errorf("yt-dlp download failed: %w", err)
		}
		return nil, fmt.Errorf("yt-dlp download failed: %w: %s", err, errText)
	}

	filePath, err := parseDownloadedFilePath(output)
	if err != nil {
		return nil, err
	}

	info, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("stat downloaded file: %w", err)
	}

	mediaProbe, err := ffprobe.ProbeMediaFile(filePath, 5*time.Second)
	if err != nil {
		return nil, fmt.Errorf("probe downloaded file: %w", err)
	}
	if err := validateProbeDuration(mediaProbe, c.MaxDurationSecs); err != nil {
		return nil, err
	}

	contentType := mime.TypeByExtension(strings.ToLower(filepath.Ext(filePath)))
	detectedMIME := detectMIMEFromProbe(mediaProbe, contentType)

	return &fetch.DownloadResult{
		Path:         filePath,
		Filename:     filepath.Base(filePath),
		Size:         info.Size(),
		ContentType:  contentType,
		DetectedMIME: detectedMIME,
	}, nil
}
