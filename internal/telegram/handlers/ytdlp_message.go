package handlers

import (
	"fmt"

	"github.com/paintingpromisesss/cobalt_bot/internal/domain/picker"
	"github.com/paintingpromisesss/cobalt_bot/internal/telegram/presenter"
	usecasepicker "github.com/paintingpromisesss/cobalt_bot/internal/usecase/picker"
	tele "gopkg.in/telebot.v4"
)

func (h *Handler) initYtDLPPicker(c tele.Context, statusMsg *tele.Message, userID int64, data *picker.YtDLPInitData) error {
	result, err := h.pickerService.InitYtDLP(usecasepicker.InitYtDLPInput{
		UserID: userID,
		Data:   *data,
	})
	if err != nil {
		return err
	}
	if result.View == nil {
		return fmt.Errorf("yt-dlp picker view is nil")
	}

	return h.renderYtDLPPicker(c, statusMsg, result.SessionID, result.View)
}

func (h *Handler) renderYtDLPPicker(c tele.Context, statusMsg *tele.Message, sessionID string, pickerView *picker.YtDLPView) error {
	markup, message := presenter.BuildYtDLPPickerMessage(sessionID, pickerView)
	_, err := c.Bot().Edit(statusMsg, message, &tele.SendOptions{ReplyMarkup: markup})
	return err
}

func (h *Handler) renderYtDLPConfirmation(c tele.Context, statusMsg *tele.Message, sessionID string, option picker.YtDLPOption) error {
	markup, message := presenter.BuildYtDLPConfirmationMessage(sessionID, option)
	_, err := c.Bot().Edit(statusMsg, message, &tele.SendOptions{ReplyMarkup: markup})
	return err
}
