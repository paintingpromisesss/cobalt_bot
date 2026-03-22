package ffprobe

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

type MediaStream struct {
	CodecType string
	CodecName string
	Width     int
	Height    int
	Duration  string
}

type MediaProbe struct {
	FormatDuration string
	Streams        []MediaStream
}

type ffprobeOutput struct {
	Streams []ffprobeStream `json:"streams"`
	Format  ffprobeFormat   `json:"format"`
}

type ffprobeStream struct {
	CodecType string `json:"codec_type"`
	CodecName string `json:"codec_name"`
	Width     int    `json:"width"`
	Height    int    `json:"height"`
	Duration  string `json:"duration"`
}

type ffprobeFormat struct {
	Duration string `json:"duration"`
}

func ProbeMediaFile(filePath string, timeout time.Duration) (MediaProbe, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "ffprobe",
		"-v", "error",
		"-show_format",
		"-show_streams",
		"-print_format", "json",
		filePath,
	)

	output, err := cmd.Output()
	if err != nil {
		return MediaProbe{}, fmt.Errorf("run ffprobe: %w", err)
	}

	var parsed ffprobeOutput
	if err := json.Unmarshal(output, &parsed); err != nil {
		return MediaProbe{}, fmt.Errorf("decode ffprobe output: %w", err)
	}

	probe := MediaProbe{
		FormatDuration: strings.TrimSpace(parsed.Format.Duration),
		Streams:        make([]MediaStream, 0, len(parsed.Streams)),
	}

	for _, stream := range parsed.Streams {
		probe.Streams = append(probe.Streams, MediaStream{
			CodecType: strings.TrimSpace(stream.CodecType),
			CodecName: strings.TrimSpace(stream.CodecName),
			Width:     stream.Width,
			Height:    stream.Height,
			Duration:  strings.TrimSpace(stream.Duration),
		})
	}

	return probe, nil
}
