package media

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	ffprobe "github.com/paintingpromisesss/cobalt_bot/internal/adapters/ffprobe"
	tele "gopkg.in/telebot.v4"
)

func applyVideoMetadata(video *tele.Video, mediaProbe ffprobe.MediaProbe) error {
	if video == nil {
		return fmt.Errorf("video is nil")
	}

	for _, stream := range mediaProbe.Streams {
		if stream.CodecType != "video" {
			continue
		}

		video.Width = stream.Width
		video.Height = stream.Height

		duration, err := parseDurationSeconds(stream.Duration)
		if err != nil {
			duration, err = parseDurationSeconds(mediaProbe.FormatDuration)
		}
		if err != nil {
			return fmt.Errorf("parse video duration: %w", err)
		}

		video.Duration = duration
		return nil
	}

	return fmt.Errorf("video stream not found")
}

func parseDurationSeconds(raw string) (int, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return 0, fmt.Errorf("empty duration")
	}

	seconds, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, err
	}
	if seconds <= 0 {
		return 0, fmt.Errorf("non-positive duration")
	}

	return int(math.Round(seconds)), nil
}
