package handlers

import (
	"context"
	"errors"
	"fmt"

	"github.com/paintingpromisesss/cobalt_bot/internal/adapters/fetch"
	"github.com/paintingpromisesss/cobalt_bot/internal/adapters/ytdlp"
	"github.com/paintingpromisesss/cobalt_bot/internal/domain/picker"
	"github.com/paintingpromisesss/cobalt_bot/internal/telegram"
	"go.uber.org/zap"
	tele "gopkg.in/telebot.v4"
)

func formatDuration(duration int) string {
	totalSeconds := duration
	hours := totalSeconds / 3600
	minutes := (totalSeconds % 3600) / 60
	seconds := totalSeconds % 60

	if hours > 0 {
		return fmt.Sprintf("%d:%02d:%02d", hours, minutes, seconds)
	}
	return fmt.Sprintf("%d:%02d", minutes, seconds)
}

func formatDurationLimit(seconds int) string {
	if seconds <= 0 {
		return formatDuration(0)
	}
	return formatDuration(seconds)
}

func (h *Handler) handleStatusMessageError(c tele.Context, statusMsg *tele.Message, userID int64, username string, err error) error {
	if _, editErr := c.Bot().Edit(statusMsg, h.pickerErrorToText(err)); editErr != nil {
		h.logger.Error(
			"failed to edit status message with error",
			zap.Int64("user_id", userID),
			zap.String("username", username),
			zap.Error(editErr),
		)
		return editErr
	}

	return telegram.MarkHandled(err)
}

func (h *Handler) handlePickerError(c tele.Context, statusMsg *tele.Message, err error) error {
	switch {
	case errors.Is(err, picker.ErrSessionExpired):
		_, err := c.Bot().Edit(statusMsg, "Время сессии истекло. Пожалуйста, попробуйте отправить ссылку заново.")
		return err
	case errors.Is(err, picker.ErrNoOptionsSelected):
		_, err := c.Bot().Edit(statusMsg, "Вы не выбрали ни одного объекта для загрузки. Пожалуйста, выберите хотя бы один и попробуйте снова.")
		return err
	default:
		_, err := c.Bot().Edit(statusMsg, h.pickerErrorToText(err))
		return err
	}
}

func (h *Handler) handlePickerCallbackError(c tele.Context, statusMsg *tele.Message, err error) error {
	if editErr := h.handlePickerError(c, statusMsg, err); editErr != nil {
		return editErr
	}

	return telegram.MarkHandled(err)
}

func (h *Handler) pickerErrorToText(err error) string {
	errorText := "Произошла ошибка при обработке вашего запроса: " + err.Error()
	switch {
	case errors.Is(err, context.DeadlineExceeded):
		errorText = "Не удалось завершить обработку вовремя. Попробуйте еще раз."
	case errors.Is(err, fetch.ErrFileTooLarge):
		errorText = "Файл слишком большой для отправки."
	case errors.Is(err, fetch.ErrEmptyFile):
		errorText = "Скачанный файл оказался пустым. Попробуйте повторить позже."
	case errors.Is(err, ytdlp.ErrMediaDurationTooLong):
		errorText = "Продолжительность медиафайла превышает допустимый лимит: " + formatDurationLimit(h.maxMediaDurationSecs) + "."
	}

	return errorText
}
