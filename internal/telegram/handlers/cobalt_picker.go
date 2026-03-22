package handlers

import (
	"context"

	"github.com/paintingpromisesss/cobalt_bot/internal/domain/picker"
	"github.com/paintingpromisesss/cobalt_bot/internal/telegram/presenter"
	usecasepicker "github.com/paintingpromisesss/cobalt_bot/internal/usecase/picker"
	tele "gopkg.in/telebot.v4"
)

func (h *Handler) handleCobaltPickerCallback(c tele.Context) error {
	h.respondToCallback(c)

	userID := c.Sender().ID
	statusMsg := c.Message()

	action, sessionID, optionIdx, err := presenter.ParseCobaltPickerCallbackData(c.Data())
	if err != nil {
		return h.handleInvalidCallbackData(c, userID, c.Data(), err)
	}

	switch action {
	case picker.CobaltActionToggle:
		result, err := h.pickerService.HandleCobalt(usecasepicker.CobaltInput{
			Action:    picker.CobaltActionToggle,
			SessionID: sessionID,
			UserID:    userID,
			OptionIdx: optionIdx,
		})
		if err != nil {
			return h.handlePickerCallbackError(c, statusMsg, err)
		}
		return h.renderCobaltPicker(c, statusMsg, sessionID, result.View)
	case picker.CobaltActionSelectAll:
		result, err := h.pickerService.HandleCobalt(usecasepicker.CobaltInput{
			Action:    picker.CobaltActionSelectAll,
			SessionID: sessionID,
			UserID:    userID,
		})
		if err != nil {
			return h.handlePickerCallbackError(c, statusMsg, err)
		}
		return h.renderCobaltPicker(c, statusMsg, sessionID, result.View)
	case picker.CobaltActionClearAll:
		result, err := h.pickerService.HandleCobalt(usecasepicker.CobaltInput{
			Action:    picker.CobaltActionClearAll,
			SessionID: sessionID,
			UserID:    userID,
		})
		if err != nil {
			return h.handlePickerCallbackError(c, statusMsg, err)
		}
		return h.renderCobaltPicker(c, statusMsg, sessionID, result.View)
	case picker.CobaltActionDownload:
		result, err := h.pickerService.HandleCobalt(usecasepicker.CobaltInput{
			Action:    picker.CobaltActionDownload,
			SessionID: sessionID,
			UserID:    userID,
		})
		if err != nil {
			return h.handlePickerCallbackError(c, statusMsg, err)
		}

		err = h.runDownloadJob(userID, func(downloadCtx context.Context) error {
			return h.mediaService.SendCobaltOptions(c, statusMsg, downloadCtx, userID, c.Recipient(), result.Options)
		})

		if err != nil {
			return h.handlePickerCallbackError(c, statusMsg, err)
		}
		return nil
	case picker.CobaltActionCancel:
		_, err := h.pickerService.HandleCobalt(usecasepicker.CobaltInput{
			Action:    picker.CobaltActionCancel,
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
