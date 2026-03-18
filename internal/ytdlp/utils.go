package ytdlp

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/paintingpromisesss/cobalt_bot/internal/domain/media"
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

func (c *Client) IdentifyYoutubeURL(url string) (bool, media.YouTubeContentKind) {
	lowerURL := strings.ToLower(strings.TrimSpace(url))
	if strings.Contains(lowerURL, "youtube.com/") || strings.Contains(lowerURL, "youtu.be/") {
		if strings.Contains(lowerURL, "music") {
			return true, media.YouTubeMusic
		}
		if strings.Contains(lowerURL, "shorts") {
			return true, media.YouTubeShorts
		}
		return true, media.YouTubeVideo
	}
	return false, media.YouTubeOther
}

func (c *Client) buildDownloadArgs(url string, formatID string, selectedFormat *media.DownloadFormat) []string {
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

func buildDownloadPostProcessArgs(formatID string, selectedFormat *media.DownloadFormat) []string {
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

func (f Format) GetDisplayName(audioFormat, videoFormat *Format) string {
	if audioFormat != nil && videoFormat != nil {
		return fmt.Sprintf(
			"%dx%d [%s] [%s] + %dkbps [%s] [%s] (merged)",
			videoFormat.Width,
			videoFormat.Height,
			formatCodecLabel(videoFormat.VCodec),
			videoFormat.FormatID,
			audioFormat.GetRoundedABR(),
			formatCodecLabel(audioFormat.ACodec),
			audioFormat.FormatID,
		)
	}

	if f.IsAudio() && !f.IsVideo() {
		return fmt.Sprintf("%dkbps [%s] [%s]", f.GetRoundedABR(), formatCodecLabel(f.ACodec), f.FormatID)
	}
	if f.IsVideo() && !f.IsAudio() {
		return fmt.Sprintf("%dx%d [%s] [%s]", f.Width, f.Height, formatCodecLabel(f.VCodec), f.FormatID)
	}

	if f.IsAudio() && f.IsVideo() {
		return fmt.Sprintf(
			"%dx%d [%s] + %dkbps [%s] (muxed) [%s]",
			f.Width,
			f.Height,
			formatCodecLabel(f.VCodec),
			f.GetRoundedABR(),
			formatCodecLabel(f.ACodec),
			f.FormatID,
		)
	}

	return f.FormatID
}

func formatCodecLabel(codec string) string {
	value := strings.TrimSpace(codec)
	if value == "" || value == "none" {
		return "unknown"
	}

	if idx := strings.Index(value, "."); idx > 0 {
		return value[:idx]
	}

	return value
}
