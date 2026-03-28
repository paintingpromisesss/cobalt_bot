package media

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	ffprobe "github.com/paintingpromisesss/cobalt_bot/internal/adapters/ffprobe"
	"go.uber.org/zap"
	tele "gopkg.in/telebot.v4"
)

type Sender struct {
	log            *zap.Logger
	ffprobeTimeout time.Duration
	ffmpegTimeout  time.Duration
	localFileMode  bool
}

func NewSender(log *zap.Logger, ffprobeTimeout time.Duration, ffmpegTimeout time.Duration, localFileMode bool) *Sender {
	return &Sender{
		log:            log,
		ffprobeTimeout: ffprobeTimeout,
		ffmpegTimeout:  ffmpegTimeout,
		localFileMode:  localFileMode,
	}
}

func (s *Sender) SendFile(c tele.Context, filePath, fileName, detectedMIME string, recipient tele.Recipient) error {
	if c == nil {
		return errors.New("telegram context is nil")
	}
	if strings.TrimSpace(filePath) == "" {
		return errors.New("file path is required")
	}
	if strings.TrimSpace(fileName) == "" {
		return errors.New("file name is required")
	}

	file, absPath, err := s.resolveTelegramFile(filePath)
	if err != nil {
		return err
	}

	media, cleanup := s.buildMedia(file, absPath, fileName, detectedMIME)
	defer cleanup()

	if _, err := c.Bot().Send(recipient, media); err != nil {
		return fmt.Errorf("send file to telegram: %w", err)
	}

	return nil
}

func (s *Sender) buildMedia(file tele.File, filePath, fileName, detectedMIME string) (any, func()) {
	mime := strings.TrimSpace(strings.ToLower(detectedMIME))
	cleanup := func() {}

	switch {
	case strings.HasPrefix(mime, "image/"):
		return &tele.Photo{File: file}, cleanup
	case strings.HasPrefix(mime, "video/"):
		video := &tele.Video{
			File:      file,
			FileName:  fileName,
			MIME:      detectedMIME,
			Streaming: true,
		}

		mediaProbe, err := ffprobe.ProbeMediaFile(filePath, s.ffprobeTimeout)
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

func (s *Sender) buildAlbumItem(filePath, fileName, detectedMIME string) (tele.Inputtable, error) {
	file, _, err := s.resolveTelegramFile(filePath)
	if err != nil {
		return nil, err
	}

	mime := strings.TrimSpace(strings.ToLower(detectedMIME))

	switch {
	case strings.HasPrefix(mime, "image/"):
		return &tele.Photo{File: file}, nil
	case strings.HasPrefix(mime, "video/"):
		return &tele.Video{File: file, FileName: fileName, MIME: detectedMIME, Streaming: true}, nil
	default:
		return &tele.Document{File: file, FileName: fileName, MIME: detectedMIME}, nil
	}
}

func (s *Sender) resolveTelegramFile(filePath string) (tele.File, string, error) {
	if strings.TrimSpace(filePath) == "" {
		return tele.File{}, "", errors.New("file path is required")
	}

	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return tele.File{}, "", fmt.Errorf("resolve absolute file path: %w", err)
	}
	if !filepath.IsAbs(absPath) {
		return tele.File{}, "", fmt.Errorf("file path must be absolute: %s", absPath)
	}

	info, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			return tele.File{}, "", fmt.Errorf("file does not exist: %s", absPath)
		}
		return tele.File{}, "", fmt.Errorf("stat file: %w", err)
	}
	if info.IsDir() {
		return tele.File{}, "", fmt.Errorf("file path points to a directory: %s", absPath)
	}

	if !s.localFileMode {
		return tele.FromDisk(absPath), absPath, nil
	}

	return tele.FromURL(localFileURI(absPath)), absPath, nil
}

func localFileURI(filePath string) string {
	return (&url.URL{
		Scheme: "file",
		Path:   filepath.ToSlash(filePath),
	}).String()
}
