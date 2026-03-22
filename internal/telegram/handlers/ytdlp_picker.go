package handlers

import (
	"context"

	"github.com/paintingpromisesss/cobalt_bot/internal/domain/picker"
	"github.com/paintingpromisesss/cobalt_bot/internal/telegram/presenter"
	usecasepicker "github.com/paintingpromisesss/cobalt_bot/internal/usecase/picker"
	tele "gopkg.in/telebot.v4"
)

func (h *Handler) handleYtDLPPickerCallback(c tele.Context) error {
	h.respondToCallback(c)

	userID := c.Sender().ID
	statusMsg := c.Message()

	action, sessionID, tab, optionIdx, err := presenter.ParseYtDLPPickerCallbackData(c.Data())
	if err != nil {
		return h.handleInvalidCallbackData(c, userID, c.Data(), err)
	}

	switch action {
	case picker.YtDLPActionTab:
		result, err := h.pickerService.HandleYtDLP(usecasepicker.YtDLPInput{
			Action:    picker.YtDLPActionTab,
			SessionID: sessionID,
			UserID:    userID,
			Tab:       tab,
		})
		if err != nil {
			return h.handlePickerCallbackError(c, statusMsg, err)
		}
		return h.renderYtDLPPicker(c, statusMsg, sessionID, result.View)
	case picker.YtDLPActionChoose:
		result, err := h.pickerService.HandleYtDLP(usecasepicker.YtDLPInput{
			Action:    picker.YtDLPActionChoose,
			SessionID: sessionID,
			UserID:    userID,
			OptionIdx: optionIdx,
		})
		if err != nil {
			return h.handlePickerCallbackError(c, statusMsg, err)
		}
		return h.renderYtDLPConfirmation(c, statusMsg, sessionID, *result.Option)
	case picker.YtDLPActionDownload:
		result, err := h.pickerService.HandleYtDLP(usecasepicker.YtDLPInput{
			Action:    picker.YtDLPActionDownload,
			SessionID: sessionID,
			UserID:    userID,
		})
		if err != nil {
			return h.handlePickerCallbackError(c, statusMsg, err)
		}

		err = h.runDownloadJob(userID, func(downloadCtx context.Context) error {
			return h.mediaService.SendYtDLPOption(c, downloadCtx, statusMsg, c.Recipient(), *result.Option)
		})
		if err != nil {
			return h.handlePickerCallbackError(c, statusMsg, err)
		}
		return nil
	case picker.YtDLPActionConfirmBack:
		result, err := h.pickerService.HandleYtDLP(usecasepicker.YtDLPInput{
			Action:    picker.YtDLPActionConfirmBack,
			SessionID: sessionID,
			UserID:    userID,
		})
		if err != nil {
			return h.handlePickerCallbackError(c, statusMsg, err)
		}
		return h.renderYtDLPPicker(c, statusMsg, sessionID, result.View)
	case picker.YtDLPActionBack:
		result, err := h.pickerService.HandleYtDLP(usecasepicker.YtDLPInput{
			Action:    picker.YtDLPActionBack,
			SessionID: sessionID,
			UserID:    userID,
			Tab:       tab,
		})
		if err != nil {
			return h.handlePickerCallbackError(c, statusMsg, err)
		}
		return h.renderYtDLPPicker(c, statusMsg, sessionID, result.View)
	case picker.YtDLPActionCancel:
		_, err := h.pickerService.HandleYtDLP(usecasepicker.YtDLPInput{
			Action:    picker.YtDLPActionCancel,
			SessionID: sessionID,
			UserID:    userID,
		})
		if err != nil {
			return h.handlePickerCallbackError(c, statusMsg, err)
		}
		_, err = c.Bot().Edit(statusMsg, "Сессия выбора отменена. Если хотите скачать что-то ещё, просто отправьте ссылку.")
		return err
	default:
		_, err := c.Bot().Edit(statusMsg, "Неизвестное действие. Пожалуйста, попробуйте снова.")
		return err
	}
}
