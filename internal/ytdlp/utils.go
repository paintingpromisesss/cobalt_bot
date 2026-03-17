package ytdlp

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
)

type YoutubeURLType string

const (
	YoutubeVideo  YoutubeURLType = "video"
	YoutubeMusic  YoutubeURLType = "music"
	YoutubeShorts YoutubeURLType = "shorts"
)

func (c *Client) buildGetMetadataArgs(url string, ClientType *YtDLPClient) []string {
	args := []string{"-J", "--skip-download"}

	if !c.PlaylistAvailable {
		args = append(args, "--no-playlist")
	}

	if c.MaxDurationSecs > 0 {
		if !c.CurrentlyLiveAvailable {
			args = append(args, "--match-filter", "duration <= "+fmt.Sprint(c.MaxDurationSecs)+" & !is_live")
		} else {
			args = append(args, "--match-filter", "duration <= "+fmt.Sprint(c.MaxDurationSecs))
		}
	}

	if c.MaxFileBytes > 0 {
		args = append(args, "--max-filesize", fmt.Sprint(c.MaxFileBytes))
	}

	if ClientType != nil {
		args = append(args, "--extractor-args", fmt.Sprintf("youtube:player_client=%s", *ClientType))
	}

	args = append(args, url)

	return args
}

func (c *Client) IdentifyYoutubeURL(url string) (bool, YoutubeURLType) {
	lowerURL := strings.ToLower(strings.TrimSpace(url))
	if strings.Contains(lowerURL, "youtube.com/") || strings.Contains(lowerURL, "youtu.be/") {
		if strings.Contains(lowerURL, "music") {
			return true, YoutubeMusic
		}
		if strings.Contains(lowerURL, "shorts") {
			return true, YoutubeShorts
		}
		return true, YoutubeVideo
	}
	return false, "other"
}

func (c *Client) buildDownloadArgs(url string, formatID string, selectedFormat *Format) []string {
	args := []string{
		"-f", formatID,
		"-P", "home:" + c.tempDir,
		"-P", "temp:" + c.tempDir,
		"-o", "%(title)s [%(id)s] [%(format_id)s].%(ext)s",
	}

	if !c.PlaylistAvailable {
		args = append(args, "--no-playlist")
	}

	args = append(args, buildDownloadPostProcessArgs(formatID, selectedFormat)...)

	if c.MaxDurationSecs > 0 {
		if !c.CurrentlyLiveAvailable {
			args = append(args, "--match-filter", "duration <= "+fmt.Sprint(c.MaxDurationSecs)+" & !is_live")
		} else {
			args = append(args, "--match-filter", "duration <= "+fmt.Sprint(c.MaxDurationSecs))
		}
	}

	if c.MaxFileBytes > 0 {
		args = append(args, "--max-filesize", fmt.Sprint(c.MaxFileBytes))
	}

	if c.ClientType != nil {
		args = append(args, "--extractor-args", fmt.Sprintf("youtube:player_client=%s", *c.ClientType))
	}

	args = append(args, url)

	args = append(args, "--print", "after_move:filepath")

	return args
}

func buildDownloadPostProcessArgs(formatID string, selectedFormat *Format) []string {
	switch {
	case strings.Contains(formatID, "+"):
		return []string{"--merge-output-format", "mp4"}
	case selectedFormat != nil && selectedFormat.IsVideo():
		return []string{"--remux-video", "mp4"}
	case selectedFormat != nil && selectedFormat.IsAudio() && !selectedFormat.IsVideo():
		return []string{"--extract-audio", "--audio-format", "mp3"}
	default:
		return nil
	}
}

func parseDownloadedFilePath(output []byte) (string, error) {
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}
		return filepath.Clean(line), nil
	}
	return "", errors.New("yt-dlp did not return downloaded filepath")
}
