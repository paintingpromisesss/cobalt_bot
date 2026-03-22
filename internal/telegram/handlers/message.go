package handlers

import (
	"context"
	"time"

	usecasedownload "github.com/paintingpromisesss/cobalt_bot/internal/usecase/download"
	usecasepicker "github.com/paintingpromisesss/cobalt_bot/internal/usecase/picker"
	"go.uber.org/zap"
	tele "gopkg.in/telebot.v4"
)

func (h *Handler) handleMessage(c tele.Context) error {
	userID := c.Sender().ID
	username := c.Sender().Username
	user := c.Recipient()
	url := c.Text()
	sessionStartedAt := time.Now()

	err := h.runDownloadJob(userID, func(downloadCtx context.Context) error {
		metaCtx, cancelMeta := context.WithTimeout(downloadCtx, h.requestTimeout)
		defer cancelMeta()

		statusMsg, err := c.Bot().Send(user, "Ваш запрос принят. Получаю информацию...")
		if err != nil {
			return err
		}

		result, err := h.downloadService.Handle(metaCtx, usecasedownload.Input{
			UserID: userID,
			URL:    url,
		})
		if err != nil {
			h.logger.Error(
				"download session failed",
				zap.Int64("user_id", userID),
				zap.String("username", username),
				zap.Duration("session_duration", time.Since(sessionStartedAt)),
				zap.Error(err),
			)
			return c.Send(h.pickerErrorToText(err))
		}

		if _, ok := result.(usecasedownload.InvalidURLResult); ok {
			h.logger.Warn(
				"user sent invalid url",
				zap.Int64("user_id", userID),
				zap.String("username", username),
				zap.String("input", url),
			)
			return editMessageText(c, statusMsg, "Неправильная ссылка. Пожалуйста, проверьте URL и попробуйте снова.")
		}

		url = result.NormalizedURL()

		h.logger.Info(
			"user started download session",
			zap.Int64("user_id", userID),
			zap.String("username", username),
			zap.String("url", url),
		)

		fields := []zap.Field{
			zap.Int64("user_id", userID),
			zap.String("username", username),
			zap.Duration("session_duration", time.Since(sessionStartedAt)),
		}

		switch r := result.(type) {
		case usecasedownload.CobaltDirectResult:
			if err := h.mediaService.SendCobaltSingle(c, downloadCtx, statusMsg, user, userID, url, r.File); err != nil {
				return h.handleStatusMessageError(c, statusMsg, userID, username, err)
			}
		case usecasedownload.CobaltPickerResult:
			if err := h.initCobaltPicker(c, statusMsg, usecasepicker.InitCobaltInput{
				UserID: userID,
				Data:   r.Data,
			}); err != nil {
				return h.handleStatusMessageError(c, statusMsg, userID, username, err)
			}
			h.logger.Info("download session awaiting selection", fields...)
		case usecasedownload.YtDLPPickerResult:
			if err := h.initYtDLPPicker(c, statusMsg, userID, &r.Data); err != nil {
				return h.handleStatusMessageError(c, statusMsg, userID, username, err)
			}
			h.logger.Info("download session awaiting selection", fields...)
		case usecasedownload.YtDLPDirectResult:
			if err := h.mediaService.SendYtDLPOption(c, downloadCtx, statusMsg, user, r.Option); err != nil {
				return h.handleStatusMessageError(c, statusMsg, userID, username, err)
			}
		default:
			return editMessageText(c, statusMsg, "Не удалось определить сценарий загрузки.")
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}
