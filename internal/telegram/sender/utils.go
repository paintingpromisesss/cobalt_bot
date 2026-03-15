package sender

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type mergeMode string

const (
	mergeModeCopy      mergeMode = "copy"
	mergeModeTranscode mergeMode = "transcode"

	mergedMediaMIME = "video/mp4"
)

var (
	errVideoStreamNotFound = errors.New("video stream not found")
	errAudioStreamNotFound = errors.New("audio stream not found")
)

type mergedMediaMetadata struct {
	Size       int64
	Width      int
	Height     int
	Duration   int
	MIME       string
	VideoCodec string
	AudioCodec string
}

type mediaProbe struct {
	FormatDuration string
	Streams        []mediaStream
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

type mediaStream struct {
	CodecType string
	CodecName string
	Width     int
	Height    int
	Duration  string
}

type ffprobeFormat struct {
	Duration string `json:"duration"`
}

func (s *FileSender) MergeStreamsToStreamableMP4(videoPath, audioPath string) (string, mergedMediaMetadata, error) {
	videoProbe, err := probeMediaFile(videoPath, s.ffprobeTimeout)
	if err != nil {
		return "", mergedMediaMetadata{}, fmt.Errorf("probe video file: %w", err)
	}

	audioProbe, err := probeMediaFile(audioPath, s.ffprobeTimeout)
	if err != nil {
		return "", mergedMediaMetadata{}, fmt.Errorf("probe audio file: %w", err)
	}

	mode, err := decideMergeMode(videoProbe, audioProbe)
	if err != nil {
		return "", mergedMediaMetadata{}, err
	}

	outputPath, err := createTempMP4("cobalt-merged-*.mp4")
	if err != nil {
		return "", mergedMediaMetadata{}, err
	}

	args := []string{
		"-y",
		"-i", videoPath,
		"-i", audioPath,
		"-map", "0:v:0",
		"-map", "1:a:0",
	}

	switch mode {
	case mergeModeCopy:
		args = append(args, "-c", "copy")
	case mergeModeTranscode:
		args = append(args,
			"-c:v", "libx264",
			"-pix_fmt", "yuv420p",
			"-c:a", "aac",
		)
	default:
		cleanupFile(outputPath)
		return "", mergedMediaMetadata{}, fmt.Errorf("unsupported merge mode %q", mode)
	}

	args = append(args,
		"-movflags", "+faststart",
		"-f", "mp4",
		outputPath,
	)

	if err := runFFmpeg(s.ffmpegTimeout, outputPath, args...); err != nil {
		return "", mergedMediaMetadata{}, err
	}

	mergedProbe, err := probeMediaFile(outputPath, s.ffprobeTimeout)
	if err != nil {
		cleanupFile(outputPath)
		return "", mergedMediaMetadata{}, fmt.Errorf("probe merged file: %w", err)
	}

	info, err := os.Stat(outputPath)
	if err != nil {
		cleanupFile(outputPath)
		return "", mergedMediaMetadata{}, fmt.Errorf("stat merged file: %w", err)
	}

	metadata, err := buildMergedMediaMetadata(info.Size(), mergedProbe)
	if err != nil {
		cleanupFile(outputPath)
		return "", mergedMediaMetadata{}, err
	}

	return filepath.Clean(outputPath), metadata, nil
}

func probeMediaFile(filePath string, timeout time.Duration) (mediaProbe, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "ffprobe",
		"-v", "error",
		"-show_streams",
		"-print_format", "json",
		filePath,
	)

	output, err := cmd.Output()
	if err != nil {
		return mediaProbe{}, fmt.Errorf("run ffprobe: %w", err)
	}

	var parsed ffprobeOutput
	if err := json.Unmarshal(output, &parsed); err != nil {
		return mediaProbe{}, fmt.Errorf("decode ffprobe output: %w", err)
	}

	probe := mediaProbe{
		FormatDuration: strings.TrimSpace(parsed.Format.Duration),
		Streams:        make([]mediaStream, 0, len(parsed.Streams)),
	}

	for _, stream := range parsed.Streams {
		probe.Streams = append(probe.Streams, mediaStream{
			CodecType: strings.TrimSpace(stream.CodecType),
			CodecName: strings.TrimSpace(stream.CodecName),
			Width:     stream.Width,
			Height:    stream.Height,
			Duration:  strings.TrimSpace(stream.Duration),
		})
	}

	return probe, nil
}

func decideMergeMode(videoProbe, audioProbe mediaProbe) (mergeMode, error) {
	videoStream, ok := videoProbe.firstStream("video")
	if !ok {
		return "", errVideoStreamNotFound
	}

	audioStream, ok := audioProbe.firstStream("audio")
	if !ok {
		return "", errAudioStreamNotFound
	}

	if isTelegramSafeVideoCodec(videoStream.CodecName) && isTelegramSafeAudioCodec(audioStream.CodecName) {
		return mergeModeCopy, nil
	}

	return mergeModeTranscode, nil
}

func buildMergedMediaMetadata(size int64, probe mediaProbe) (mergedMediaMetadata, error) {
	videoStream, ok := probe.firstStream("video")
	if !ok {
		return mergedMediaMetadata{}, errVideoStreamNotFound
	}

	audioStream, ok := probe.firstStream("audio")
	if !ok {
		return mergedMediaMetadata{}, errAudioStreamNotFound
	}

	duration, err := parseDurationSeconds(videoStream.Duration)
	if err != nil {
		duration, err = parseDurationSeconds(probe.FormatDuration)
		if err != nil {
			return mergedMediaMetadata{}, fmt.Errorf("extract merged duration: %w", err)
		}
	}

	return mergedMediaMetadata{
		Size:       size,
		Width:      videoStream.Width,
		Height:     videoStream.Height,
		Duration:   duration,
		VideoCodec: videoStream.CodecName,
		AudioCodec: audioStream.CodecName,
		MIME:       mergedMediaMIME,
	}, nil
}

func createTempMP4(pattern string) (string, error) {
	outputFile, err := os.CreateTemp("", pattern)
	if err != nil {
		return "", fmt.Errorf("create temp mp4: %w", err)
	}

	outputPath := outputFile.Name()
	if err := outputFile.Close(); err != nil {
		cleanupFile(outputPath)
		return "", fmt.Errorf("close temp mp4: %w", err)
	}

	return outputPath, nil
}

func runFFmpeg(timeout time.Duration, outputPath string, args ...string) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "ffmpeg", args...)
	if output, err := cmd.CombinedOutput(); err != nil {
		cleanupFile(outputPath)
		return fmt.Errorf("run ffmpeg: %w: %s", err, strings.TrimSpace(string(output)))
	}

	info, err := os.Stat(outputPath)
	if err != nil {
		cleanupFile(outputPath)
		return fmt.Errorf("stat output file: %w", err)
	}
	if info.Size() == 0 {
		cleanupFile(outputPath)
		return fmt.Errorf("output file is empty")
	}

	return nil
}

func (p mediaProbe) firstStream(codecType string) (mediaStream, bool) {
	for _, stream := range p.Streams {
		if stream.CodecType == codecType {
			return stream, true
		}
	}
	return mediaStream{}, false
}

func isTelegramSafeVideoCodec(codec string) bool {
	return strings.EqualFold(strings.TrimSpace(codec), "h264")
}

func isTelegramSafeAudioCodec(codec string) bool {
	return strings.EqualFold(strings.TrimSpace(codec), "aac")
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

func cleanupFile(path string) {
	if strings.TrimSpace(path) == "" {
		return
	}
	_ = os.Remove(path)
}
