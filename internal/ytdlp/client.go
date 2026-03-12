package ytdlp

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
)

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

	return &metadata, nil
}
