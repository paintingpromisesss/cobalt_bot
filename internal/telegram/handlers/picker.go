package handlers

import (
	"context"

	"github.com/paintingpromisesss/cobalt_bot/internal/telegram"
	pickersession "github.com/paintingpromisesss/cobalt_bot/internal/telegram/picker_session"
	"go.uber.org/zap"
	tele "gopkg.in/telebot.v4"
)

func (h *Handler) handlePickerCallback(c tele.Context) error {
	if err := c.Respond(); err != nil {
		h.logger.Warn("failed to respond to picker callback", zap.Error(err))
	}

	userID := c.Sender().ID
	statusMsg := c.Message()

	action, sessionID, optionIdx, err := parsePickerCallbackData(c.Data())
	if err != nil {
		h.logger.Warn("failed to parse picker callback data", zap.Int64("user_id", userID), zap.String("data", c.Data()), zap.Error(err))

		if err := c.Edit("Не удалось распознать действие. Пожалуйста, попробуйте снова."); err != nil {
			return err
		}

		return telegram.MarkHandled(err)
	}

	switch action {
	case ToggleAction:
		pickerView, err := h.pickerSessionManager.ToggleCobaltPickerOption(sessionID, userID, optionIdx)
		if err != nil {
			return handlePickerCallbackError(c, statusMsg, err)
		}
		return h.renderPickerKeyboard(c, statusMsg, sessionID, &pickerView)
	case SelectAllAction:
		pickerView, err := h.pickerSessionManager.MarkAllCobaltPickerOptions(sessionID, userID, true)
		if err != nil {
			return handlePickerCallbackError(c, statusMsg, err)
		}
		return h.renderPickerKeyboard(c, statusMsg, sessionID, &pickerView)
	case ClearAllAction:
		pickerView, err := h.pickerSessionManager.MarkAllCobaltPickerOptions(sessionID, userID, false)
		if err != nil {
			return handlePickerCallbackError(c, statusMsg, err)
		}
		return h.renderPickerKeyboard(c, statusMsg, sessionID, &pickerView)
	case DownloadAction:
		options, err := h.pickerSessionManager.ConsumeSelectedCobaltOptions(sessionID, userID)
		if err != nil {
			return handlePickerCallbackError(c, statusMsg, err)
		}

		err = h.queueManager.Run(userID, func() error {
			downloadCtx, cancel := context.WithTimeout(h.appCtx, h.downloadTimeout)
			defer cancel()
			return h.DownloadAndSendSelectedOptions(c, statusMsg, downloadCtx, userID, c.Recipient(), options)
		})

		if err != nil {
			return handlePickerCallbackError(c, statusMsg, err)
		}
		return nil
	case CancelAction:
		err := h.pickerSessionManager.DeleteSession(sessionID, userID, pickersession.PickerSessionTypeCobalt)
		if err != nil {
			return handlePickerCallbackError(c, statusMsg, err)
		}
		_, err = c.Bot().Edit(statusMsg, "Сессия выбора отменена. Если хотите скачать что-то ещё, просто отправьте ссылку.")
		return err
	default:
		_, err := c.Bot().Edit(statusMsg, "Неизвестное действие. Пожалуйста, попробуйте снова.")
		return err
	}
}
