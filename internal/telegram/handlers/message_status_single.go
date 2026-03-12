package handlers

import (
	"context"
	"fmt"

	"github.com/paintingpromisesss/cobalt_bot/internal/cobalt"
	"go.uber.org/zap"
	tele "gopkg.in/telebot.v4"
)

func (h *Handler) handleMessageStatusSingle(c tele.Context, ctx context.Context, statusMsg *tele.Message, user tele.Recipient, userID int64, sourceURL string, cobaltResponse cobalt.MainResponse) error {
	if _, err := c.Bot().Edit(statusMsg, "Информация о файле получена. Имя файла: "+cobaltResponse.Filename+". Начинаю загрузку..."); err != nil {
		return err
	}

	downloadResult, err := h.downloader.Download(ctx, cobaltResponse.Url, cobaltResponse.Filename)
	if err != nil {
		h.logger.Error(
			"failed to download file",
			zap.Int64("user_id", userID),
			zap.String("source_url", sourceURL),
			zap.String("url", cobaltResponse.Url),
			zap.String("filename", cobaltResponse.Filename),
			zap.Error(err),
		)
		return err
	}

	if downloadResult.Size <= 0 {
		return fmt.Errorf("downloaded empty file: %s", downloadResult.Filename)
	}

	h.logger.Info(
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

	defer cleanupTempFile(h.logger, downloadResult.Path)

	if err := h.sender.SendFile(c, downloadResult.Path, downloadResult.Filename, downloadResult.DetectedMIME, user); err != nil {
		return err
	}

	h.logger.Info(
		"file sent successfully",
		zap.Int64("user_id", userID),
		zap.String("filename", downloadResult.Filename),
		zap.String("detected_mime", downloadResult.DetectedMIME),
		zap.Int64("size", downloadResult.Size),
	)

	return nil
}
