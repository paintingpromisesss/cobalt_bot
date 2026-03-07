package handlers

import (
	"fmt"
	"strconv"

	"github.com/paintingpromisesss/cobalt_bot/internal/cobalt"
	pickersession "github.com/paintingpromisesss/cobalt_bot/internal/telegram/picker_session"
	tele "gopkg.in/telebot.v4"
)

const (
	NeutralIndicator    = "⬜"
	SelectedIndicator   = "✅"
	UnselectedIndicator = "❌"
	DownloadIndicator   = "⬇️"

	ToggleAction    = "toggle"
	SelectAllAction = "select_all"
	ClearAllAction  = "clear_all"
	DownloadAction  = "download"

	PickerButtonUnique = "picker_button"
)

// handleMessageStatusPicker реализует обработку статуса Picker от Cobalt, который возвращает список объектов для скачивания.
func (h *Handler) handleMessageStatusPicker(c tele.Context, statusMsg *tele.Message, userID int64, cobaltResponse cobalt.MainResponse) error {
	pickerSessionID := h.pickerSessionManager.CreateSession(userID, cobaltResponse.Picker)
	pickerView, err := h.pickerSessionManager.GetPickerView(pickerSessionID, userID)
	if err != nil {
		return err
	}

	return h.renderPickerKeyboard(c, statusMsg, pickerSessionID, &pickerView)
}

func (h *Handler) renderPickerKeyboard(c tele.Context, statusMsg *tele.Message, sessionID string, pickerView *pickersession.PickerView) error {
	markup, message := buildPickerMessage(sessionID, pickerView)
	_, err := c.Bot().Edit(statusMsg, message, &tele.SendOptions{ReplyMarkup: markup})
	return err
}

func buildPickerMessage(sessionID string, pickerView *pickersession.PickerView) (*tele.ReplyMarkup, string) {
	markup := &tele.ReplyMarkup{}
	total := len(pickerView.Options)
	rows := make([]tele.Row, 0, total+2)
	selected := 0

	for i, option := range pickerView.Options {
		indicator := NeutralIndicator
		if option.Selected {
			indicator = SelectedIndicator
			selected++
		}
		payload := encodePickerCallbackData(ToggleAction, sessionID, i)
		rows = append(rows, markup.Row(markup.Data(indicator+" "+option.Label, PickerButtonUnique, payload)))

	}

	if selected > 0 {
		rows = append(rows, markup.Row(
			markup.Data(UnselectedIndicator+" Очистить выбор", PickerButtonUnique, encodePickerCallbackData(ClearAllAction, sessionID, -1)),
			markup.Data(DownloadIndicator+" Скачать", PickerButtonUnique, encodePickerCallbackData(DownloadAction, sessionID, -1)),
		))
	} else {
		rows = append(rows, markup.Row(markup.Data(SelectedIndicator+" Выбрать все", PickerButtonUnique, encodePickerCallbackData(SelectAllAction, sessionID, -1))))
	}

	markup.Inline(rows...)

	message := fmt.Sprintf("Найдено %d объектов. Выбрано: %d.\n Отметьте нужные и нажмите «Скачать».", total, selected)
	return markup, message
}

func encodePickerCallbackData(action, sessionID string, optionIdx int) string {
	if optionIdx >= 0 {
		return action + ":" + sessionID + ":" + strconv.Itoa(optionIdx)
	}
	return action + ":" + sessionID
}
