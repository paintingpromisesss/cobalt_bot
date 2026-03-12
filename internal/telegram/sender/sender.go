package sender

import (
	"errors"
	"fmt"
	"strings"
	"time"

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

// TODO: fix this
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
		streamablePath, err := remuxStreamableMP4(filePath, s.ffmpegTimeout)
		if err != nil {
			s.log.Debug("video remux failed, sending original file", zap.String("path", filePath), zap.Error(err))
			streamablePath = filePath
		} else {
			cleanup = chainCleanup(cleanup, func() {
				cleanupFile(streamablePath)
			})
			file = tele.FromDisk(streamablePath)
		}

		video := &tele.Video{
			File:      file,
			FileName:  fileName,
			MIME:      detectedMIME,
			Streaming: true,
		}
		meta, err := probeVideoMetadata(streamablePath, s.ffprobeTimeout)
		if err != nil {
			s.log.Debug("video probe failed, sending without explicit metadata", zap.String("path", streamablePath), zap.Error(err))
			return video, cleanup
		}

		video.Width = meta.Width
		video.Height = meta.Height
		video.Duration = meta.Duration

		thumbPath, err := generateVideoThumbnail(streamablePath, meta.Duration, s.ffmpegTimeout)
		if err != nil {
			s.log.Debug("video thumbnail generation failed", zap.String("path", streamablePath), zap.Error(err))
			return video, cleanup
		}

		video.Thumbnail = &tele.Photo{File: tele.FromDisk(thumbPath)}
		cleanup = chainCleanup(cleanup, func() {
			cleanupFile(thumbPath)
		})

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

func chainCleanup(cleanups ...func()) func() {
	return func() {
		for _, cleanup := range cleanups {
			if cleanup != nil {
				cleanup()
			}
		}
	}
}
