package handlers

import (
	"fmt"

	"github.com/paintingpromisesss/cobalt_bot/internal/domain/picker"
	"github.com/paintingpromisesss/cobalt_bot/internal/telegram/presenter"
	usecasepicker "github.com/paintingpromisesss/cobalt_bot/internal/usecase/picker"
	tele "gopkg.in/telebot.v4"
)

func (h *Handler) initCobaltPicker(c tele.Context, statusMsg *tele.Message, input usecasepicker.InitCobaltInput) error {
	result, err := h.pickerService.InitCobalt(input)
	if err != nil {
		return fmt.Errorf("init cobalt picker: %w", err)
	}
	if result.View == nil {
		return fmt.Errorf("cobalt picker view is nil")
	}

	return h.renderCobaltPicker(c, statusMsg, result.SessionID, result.View)
}

func (h *Handler) renderCobaltPicker(c tele.Context, statusMsg *tele.Message, sessionID string, pickerView *picker.CobaltView) error {
	markup, message := presenter.BuildCobaltPickerMessage(sessionID, pickerView)
	_, err := c.Bot().Edit(statusMsg, message, &tele.SendOptions{ReplyMarkup: markup})
	return err
}
