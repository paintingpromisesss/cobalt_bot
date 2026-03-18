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
	"strconv"
	"strings"
	"time"

	"github.com/paintingpromisesss/cobalt_bot/internal/domain/media"
	"github.com/paintingpromisesss/cobalt_bot/internal/downloader"
	"github.com/paintingpromisesss/cobalt_bot/internal/probe"
)

var ErrMediaDurationTooLong = errors.New("media duration exceeds limit")

type Client struct {
	tempDir                string
	MaxDurationSecs        int
	MaxFileBytes           int64
	CurrentlyLiveAvailable bool
	PlaylistAvailable      bool
	ClientType             *YtDLPClient
}

func NewClient(tempDir string, maxDurationSecs int, maxFileBytes int64, currentlyLiveAvailable bool, playlistAvailable bool) *Client {
	return &Client{
		tempDir:                tempDir,
		MaxDurationSecs:        maxDurationSecs,
		MaxFileBytes:           maxFileBytes,
		CurrentlyLiveAvailable: currentlyLiveAvailable,
		PlaylistAvailable:      playlistAvailable,
	}
}

func (c *Client) GetMetadata(ctx context.Context, url string) (*Metadata, error) {
	args := c.buildGetMetadataArgs(url, c.ClientType)
	cmd := exec.CommandContext(ctx, "yt-dlp", args...)

	cmd.Env = append(os.Environ(),
		"HOME="+c.tempDir,
		"XDG_CACHE_HOME="+c.tempDir,
		"TMPDIR="+c.tempDir,
		"TEMP="+c.tempDir,
		"TMP="+c.tempDir,
	)
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

func (c *Client) Download(ctx context.Context, url, formatID string, selectedFormat *media.DownloadFormat) (*downloader.DownloadResult, error) {
	args := c.buildDownloadArgs(url, formatID, selectedFormat)
	cmd := exec.CommandContext(ctx, "yt-dlp", args...)

	cmd.Env = append(os.Environ(),
		"HOME="+c.tempDir,
		"XDG_CACHE_HOME="+c.tempDir,
		"TMPDIR="+c.tempDir,
		"TEMP="+c.tempDir,
		"TMP="+c.tempDir,
	)

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

	mediaProbe, err := probe.ProbeMediaFile(filePath, 5*time.Second)
	if err != nil {
		return nil, fmt.Errorf("probe downloaded file: %w", err)
	}
	if err := validateProbeDuration(mediaProbe, c.MaxDurationSecs); err != nil {
		return nil, err
	}

	contentType := mime.TypeByExtension(strings.ToLower(filepath.Ext(filePath)))
	detectedMIME := detectMIMEFromProbe(mediaProbe, contentType)

	return &downloader.DownloadResult{
		Path:         filePath,
		Filename:     filepath.Base(filePath),
		Size:         info.Size(),
		ContentType:  contentType,
		DetectedMIME: detectedMIME,
	}, nil
}

func detectMIMEFromProbe(mediaProbe probe.MediaProbe, fallback string) string {
	if fallback == "" {
		fallback = "application/octet-stream"
	}

	hasVideo := false
	hasAudio := false
	for _, stream := range mediaProbe.Streams {
		switch stream.CodecType {
		case "video":
			hasVideo = true
		case "audio":
			hasAudio = true
		}
	}

	switch {
	case hasVideo && strings.HasPrefix(fallback, "video/"):
		return fallback
	case hasAudio && strings.HasPrefix(fallback, "audio/"):
		return fallback
	case hasVideo:
		return "video/mp4"
	case hasAudio:
		return "audio/mpeg"
	default:
		return fallback
	}
}

func validateMediaDurationSeconds(actualSeconds, maxSeconds int) error {
	if maxSeconds <= 0 || actualSeconds <= 0 {
		return nil
	}
	if actualSeconds > maxSeconds {
		return fmt.Errorf("%w: got %ds, max %ds", ErrMediaDurationTooLong, actualSeconds, maxSeconds)
	}
	return nil
}

func validateProbeDuration(mediaProbe probe.MediaProbe, maxSeconds int) error {
	if maxSeconds <= 0 {
		return nil
	}

	raw := strings.TrimSpace(mediaProbe.FormatDuration)
	if raw == "" {
		return nil
	}

	seconds, err := strconv.ParseFloat(raw, 64)
	if err != nil || seconds <= 0 {
		return nil
	}

	if seconds > float64(maxSeconds) {
		return fmt.Errorf("%w: got %.3fs, max %ds", ErrMediaDurationTooLong, seconds, maxSeconds)
	}

	return nil
}
