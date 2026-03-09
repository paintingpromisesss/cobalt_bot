package handlers

import (
	"context"
	"time"

	"go.uber.org/zap"
	tele "gopkg.in/telebot.v4"
)

func (h *Handler) handlePickerCallback(c tele.Context) error {
	if err := c.Respond(); err != nil {
		h.logger.Warn("Failed to respond to picker callback: %v", zap.Error(err))
	}

	userID := c.Sender().ID
	statusMsg := c.Message()

	action, sessionID, optionIdx, err := parsePickerCallbackData(c.Data())
	if err != nil {
		err = c.Edit("Не удалось распознать действие. Пожалуйста, попробуйте снова.")
		return err
	}

	switch action {
	case ToggleAction:
		pickerView, err := h.pickerSessionManager.TogglePickerOption(sessionID, userID, optionIdx)
		if err != nil {
			return handlePickerError(c, statusMsg, err)
		}
		return h.renderPickerKeyboard(c, statusMsg, sessionID, &pickerView)
	case SelectAllAction:
		pickerView, err := h.pickerSessionManager.MarkAllPickerOptions(sessionID, userID, true)
		if err != nil {
			return handlePickerError(c, statusMsg, err)
		}
		return h.renderPickerKeyboard(c, statusMsg, sessionID, &pickerView)
	case ClearAllAction:
		pickerView, err := h.pickerSessionManager.MarkAllPickerOptions(sessionID, userID, false)
		if err != nil {
			return handlePickerError(c, statusMsg, err)
		}
		return h.renderPickerKeyboard(c, statusMsg, sessionID, &pickerView)
	case DownloadAction:
		options, err := h.pickerSessionManager.ConsumeSelectedOptions(sessionID, userID)
		if err != nil {
			return handlePickerError(c, statusMsg, err)
		}

		err = h.queueManager.Run(userID, func() error {
			downloadCtx, cancel := context.WithTimeout(h.appCtx, 10*time.Minute)
			defer cancel()
			return h.DownloadSelectedOptions(c, statusMsg, downloadCtx, userID, options)
		})

		if err != nil {
			return handlePickerError(c, statusMsg, err)
		}
		return nil
	case CancelAction:
		err := h.pickerSessionManager.DeleteSession(sessionID, userID)
		if err != nil {
			return handlePickerError(c, statusMsg, err)
		}
		_, err = c.Bot().Edit(statusMsg, "Сессия выбора отменена. Если хотите скачать что-то ещё, просто отправьте ссылку.")
		return err
	default:
		_, err := c.Bot().Edit(statusMsg, "Неизвестное действие. Пожалуйста, попробуйте снова.")
		return err
	}
}
