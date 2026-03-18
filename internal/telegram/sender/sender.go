package sender

import (
	"errors"
	"fmt"
	"strings"
	"time"

	probe "github.com/paintingpromisesss/cobalt_bot/internal/probe"
	"go.uber.org/zap"
	tele "gopkg.in/telebot.v4"
)

type FileSender struct {
	log            *zap.Logger
	ffprobeTimeout time.Duration
	ffmpegTimeout  time.Duration
}

func NewFileSender(log *zap.Logger, ffprobeTimeout time.Duration, ffmpegTimeout time.Duration) *FileSender {
	return &FileSender{
		log:            log,
		ffprobeTimeout: ffprobeTimeout,
		ffmpegTimeout:  ffmpegTimeout,
	}
}

func (s *FileSender) SendFile(c tele.Context, filePath, fileName, detectedMIME string, recipient tele.Recipient) error {
	if c == nil {
		return errors.New("telegram context is nil")
	}
	if strings.TrimSpace(filePath) == "" {
		return errors.New("file path is required")
	}
	if strings.TrimSpace(fileName) == "" {
		return errors.New("file name is required")
	}

	media, cleanup := s.buildMedia(filePath, fileName, detectedMIME)
	defer cleanup()

	if _, err := c.Bot().Send(recipient, media); err != nil {
		return fmt.Errorf("send file to telegram: %w", err)
	}

	return nil
}

func (s *FileSender) buildMedia(filePath, fileName, detectedMIME string) (any, func()) {
	file := tele.FromDisk(filePath)
	mime := strings.TrimSpace(strings.ToLower(detectedMIME))
	cleanup := func() {}

	switch {
	case strings.HasPrefix(mime, "image/"):
		return &tele.Photo{
			File: file,
		}, cleanup
	case strings.HasPrefix(mime, "video/"):
		video := &tele.Video{
			File:      file,
			FileName:  fileName,
			MIME:      detectedMIME,
			Streaming: true,
		}

		mediaProbe, err := probe.ProbeMediaFile(filePath, s.ffprobeTimeout)
		if err != nil {
			s.log.Warn("failed to probe video metadata", zap.String("path", filePath), zap.Error(err))
		} else if err := applyVideoMetadata(video, mediaProbe); err != nil {
			s.log.Warn("failed to apply video metadata", zap.String("path", filePath), zap.Error(err))
		}

		return video, cleanup
	case strings.HasPrefix(mime, "audio/"):
		return &tele.Audio{
			File:     file,
			FileName: fileName,
			MIME:     detectedMIME,
		}, cleanup
	default:
		return &tele.Document{
			File:     file,
			FileName: fileName,
			MIME:     detectedMIME,
		}, cleanup
	}
}
