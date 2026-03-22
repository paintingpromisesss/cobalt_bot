package media

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/paintingpromisesss/cobalt_bot/internal/adapters/fetch"
	"github.com/paintingpromisesss/cobalt_bot/internal/adapters/ytdlp"
	"github.com/paintingpromisesss/cobalt_bot/internal/domain/media"
	"github.com/paintingpromisesss/cobalt_bot/internal/domain/picker"
	"go.uber.org/zap"
	tele "gopkg.in/telebot.v4"
)

const maxAlbumFiles = 10

type Service struct {
	log        *zap.Logger
	downloader *fetch.Downloader
	ytdlp      *ytdlp.Client
	sender     *Sender
}

func NewService(log *zap.Logger, downloader *fetch.Downloader, ytdlpClient *ytdlp.Client, sender *Sender) *Service {
	return &Service{
		log:        log,
		downloader: downloader,
		ytdlp:      ytdlpClient,
		sender:     sender,
	}
}

func (s *Service) SendCobaltSingle(c tele.Context, downloadCtx context.Context, statusMsg *tele.Message, user tele.Recipient, userID int64, sourceURL string, file media.RemoteFile) error {
	if _, err := c.Bot().Edit(statusMsg, "Информация о файле получена. Имя файла: "+file.Filename+". Начинаю загрузку..."); err != nil {
		return err
	}

	downloadResult, err := s.downloader.Download(downloadCtx, file.URL, file.Filename, nil)
	if err != nil {
		s.log.Error(
			"failed to download file",
			zap.Int64("user_id", userID),
			zap.String("source_url", sourceURL),
			zap.String("url", file.URL),
			zap.String("filename", file.Filename),
			zap.Error(err),
		)
		return err
	}
	defer cleanupTempFile(s.log, downloadResult.Path)

	if downloadResult.Size <= 0 {
		return fmt.Errorf("downloaded empty file: %s", downloadResult.Filename)
	}

	s.log.Info(
		"download completed",
		zap.Int64("user_id", userID),
		zap.String("path", downloadResult.Path),
		zap.String("filename", downloadResult.Filename),
		zap.Int64("size", downloadResult.Size),
		zap.String("content_type", downloadResult.ContentType),
		zap.String("detected_mime", downloadResult.DetectedMIME),
	)

	if _, err := c.Bot().Edit(statusMsg, "Файл загружен. Отправляю вам..."); err != nil {
		return err
	}

	if err := s.sender.SendFile(c, downloadResult.Path, downloadResult.Filename, downloadResult.DetectedMIME, user); err != nil {
		return err
	}

	s.log.Info(
		"file sent successfully",
		zap.Int64("user_id", userID),
		zap.String("filename", downloadResult.Filename),
		zap.String("detected_mime", downloadResult.DetectedMIME),
		zap.Int64("size", downloadResult.Size),
	)

	return nil
}

func (s *Service) SendCobaltOptions(c tele.Context, statusMsg *tele.Message, downloadCtx context.Context, userID int64, user tele.Recipient, options []picker.CobaltOption) error {
	if len(options) == 0 {
		return picker.ErrNoOptionsSelected
	}

	if _, err := c.Bot().Edit(statusMsg, fmt.Sprintf("Выбрано файлов: %d. Начинаю загрузку...", len(options))); err != nil {
		return err
	}

	downloadResults := make([]fetch.DownloadResult, 0, len(options))
	for _, option := range options {
		result, err := s.downloader.Download(downloadCtx, option.URL, option.Filename, nil)
		if err != nil {
			for _, obj := range downloadResults {
				cleanupTempFile(s.log, obj.Path)
			}
			return err
		}
		downloadResults = append(downloadResults, result)
	}
	defer func() {
		for _, obj := range downloadResults {
			cleanupTempFile(s.log, obj.Path)
		}
	}()

	for _, result := range downloadResults {
		if result.Size <= 0 {
			return fmt.Errorf("downloaded file is empty: %s", result.Filename)
		}
	}

	if len(downloadResults) == 1 {
		result := downloadResults[0]
		if _, err := c.Bot().Edit(statusMsg, "Загрузка завершена. Отправляю файл..."); err != nil {
			return err
		}
		if err := s.sender.SendFile(c, result.Path, result.Filename, result.DetectedMIME, user); err != nil {
			return err
		}
		s.log.Info(
			"download session completed",
			zap.Int64("user_id", userID),
			zap.String("username", c.Sender().Username),
			zap.Int("files_sent", 1),
		)
		return nil
	}

	if _, err := c.Bot().Edit(statusMsg, "Загрузка завершена. Отправляю файлы..."); err != nil {
		return err
	}

	for start := 0; start < len(downloadResults); start += maxAlbumFiles {
		end := start + maxAlbumFiles
		if end > len(downloadResults) {
			end = len(downloadResults)
		}

		album := make(tele.Album, 0, end-start)
		for _, result := range downloadResults[start:end] {
			album = append(album, buildAlbumItem(result.Path, result.Filename, result.DetectedMIME))
		}

		if _, err := c.Bot().SendAlbum(user, album); err != nil {
			return fmt.Errorf("failed to send album: %w", err)
		}
	}

	if _, err := c.Bot().Edit(statusMsg, fmt.Sprintf("Готово. Отправлено файлов: %d.", len(downloadResults))); err != nil {
		return err
	}
	s.log.Info(
		"download session completed",
		zap.Int64("user_id", userID),
		zap.String("username", c.Sender().Username),
		zap.Int("files_sent", len(downloadResults)),
	)
	return nil
}

func (s *Service) SendYtDLPOption(c tele.Context, downloadCtx context.Context, statusMsg *tele.Message, user tele.Recipient, option picker.YtDLPOption) error {
	if _, err := c.Bot().Edit(statusMsg, fmt.Sprintf("Начинаю загрузку формата: %s...", option.DisplayName)); err != nil {
		return err
	}

	var selectedFormat *media.DownloadFormat
	if option.Format.IsAudio() || option.Format.IsVideo() {
		selectedFormat = &option.Format
	}

	downloadResult, err := s.ytdlp.Download(downloadCtx, option.ContentURL, option.FormatID, selectedFormat)
	if err != nil {
		return err
	}
	defer cleanupTempFile(s.log, downloadResult.Path)

	if downloadResult.Size <= 0 {
		return fmt.Errorf("downloaded file is empty: %s", downloadResult.Filename)
	}

	if _, err := c.Bot().Edit(statusMsg, "Загрузка завершена. Отправляю файл..."); err != nil {
		return err
	}

	if err := s.sender.SendFile(c, downloadResult.Path, downloadResult.Filename, downloadResult.DetectedMIME, user); err != nil {
		return err
	}

	return nil
}

func cleanupTempFile(log *zap.Logger, filePath string) {
	if filePath == "" {
		return
	}
	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		log.Warn("failed to remove temp file", zap.String("path", filePath), zap.Error(err))
	}
}

func buildAlbumItem(filepath, filename, detectedMIME string) tele.Inputtable {
	file := tele.FromDisk(filepath)
	mime := strings.TrimSpace(strings.ToLower(detectedMIME))

	switch {
	case strings.HasPrefix(mime, "image/"):
		return &tele.Photo{File: file}
	case strings.HasPrefix(mime, "video/"):
		return &tele.Video{File: file, FileName: filename, MIME: detectedMIME, Streaming: true}
	default:
		return &tele.Document{File: file, FileName: filename, MIME: detectedMIME}
	}
}
