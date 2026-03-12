package ytdlp

import (
	"fmt"
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
			args = append(args, "--match-filter duration <= "+fmt.Sprint(c.MaxDurationSecs)+" & !is_live")
		} else {
			args = append(args, "--match-filter duration <= "+fmt.Sprint(c.MaxDurationSecs))
		}
	}

	if c.MaxFileBytes > 0 {
		args = append(args, "--max-filesize "+fmt.Sprint(c.MaxFileBytes))
	}

	if ClientType != nil {
		args = append(args, "--extractor-args "+fmt.Sprintf("youtube:player_client=%s", *ClientType))
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
