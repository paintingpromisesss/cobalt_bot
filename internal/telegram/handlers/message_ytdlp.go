package handlers

import (
	"context"
	"fmt"
	"strconv"

	pickersession "github.com/paintingpromisesss/cobalt_bot/internal/telegram/picker_session"
	"github.com/paintingpromisesss/cobalt_bot/internal/ytdlp"
	tele "gopkg.in/telebot.v4"
)

const (
	YtDLPActionTab      = "tab"
	YtDLPActionChoose   = "choose"
	YtDLPActionDownload = "download"
	YtDLPActionCancel   = "cancel"
	YtDLPActionBack     = "back"
)

// TODO: implement this
func (h *Handler) handleYoutubeVideoRequest(c tele.Context, downloadCtx context.Context, statusMsg *tele.Message, user tele.Recipient, userID int64, url string, meta *ytdlp.Metadata) error {
	pickerSessionID, err := h.pickerSessionManager.CreateYtDLPSession(userID, meta)
	if err != nil {
		return err
	}

	pickerView, err := h.pickerSessionManager.GetYtDLPPickerView(pickerSessionID, userID)
	if err != nil {
		return err
	}

	return h.renderYtDLPPickerKeyboard(c, statusMsg, pickerSessionID, &pickerView)
}

func (h *Handler) renderYtDLPPickerKeyboard(c tele.Context, statusMsg *tele.Message, sessionID string, pickerView *pickersession.YtDLPPickerView) error {
	var markup *tele.ReplyMarkup
	var message string
	if pickerView.ActiveTab == pickersession.YtDLPPickerTabNone {
		markup, message = buildYtDLPTabsMessage(sessionID, pickerView)
	} else {
		markup, message = buildYtDLPOptionsMessage(sessionID, pickerView)
	}

	_, err := c.Bot().Edit(statusMsg, message, &tele.SendOptions{ReplyMarkup: markup})
	return err
}

func buildYtDLPTabsMessage(sessionID string, pickerView *pickersession.YtDLPPickerView) (*tele.ReplyMarkup, string) {
	markup := &tele.ReplyMarkup{}
	total := len(pickerView.Tabs)
	rows := make([]tele.Row, 0, total+1)
	for _, tab := range pickerView.Tabs {
		payload := encodeYtDLPPickerCallbackData(YtDLPActionTab, sessionID, tab, -1)
		text := getYtDLPTabLabel(tab)

		rows = append(rows, markup.Row(markup.Data(text, YtDLPPickerButtonUnique, payload)))
	}

	payload := encodeYtDLPPickerCallbackData(YtDLPActionCancel, sessionID, pickersession.YtDLPPickerTabNone, -1)
	rows = append(rows, markup.Row(markup.Data("Отменить", YtDLPPickerButtonUnique, payload)))

	markup.Inline(rows...)

	message := fmt.Sprintf("Скачиваемый контент: %s \n Выберите опцию скачивания:", pickerView.ContentName)

	return markup, message
}

func buildYtDLPOptionsMessage(sessionID string, pickerView *pickersession.YtDLPPickerView) (*tele.ReplyMarkup, string) {
	markup := &tele.ReplyMarkup{}
	rows := make([]tele.Row, 0, len(pickerView.Options)+1)

	for i, option := range pickerView.Options {
		rows = append(rows, markup.Row(markup.Data(option.DisplayName, YtDLPPickerButtonUnique, encodeYtDLPPickerCallbackData(YtDLPActionChoose, sessionID, pickerView.ActiveTab, i))))
	}

	rows = append(rows, markup.Row(markup.Data("Назад", YtDLPPickerButtonUnique, encodeYtDLPPickerCallbackData(YtDLPActionBack, sessionID, pickersession.YtDLPPickerTabNone, -1))))

	markup.Inline(rows...)

	message := fmt.Sprintf("Выберите формат скачивания для: %s \n (тип: %s)", pickerView.ContentName, pickerView.ActiveTab)

	return markup, message
}

func (h *Handler) renderYtDLPConfirmationKeyboard(c tele.Context, statusMsg *tele.Message, sessionID string, option pickersession.YtDLPPickerOption) error {
	markup, message := buildYtDLPConfirmationMessage(sessionID, option)
	_, err := c.Bot().Edit(statusMsg, message, &tele.SendOptions{ReplyMarkup: markup})
	return err
}

func buildYtDLPConfirmationMessage(sessionID string, option pickersession.YtDLPPickerOption) (*tele.ReplyMarkup, string) {
	markup := &tele.ReplyMarkup{}
	rows := make([]tele.Row, 0, 2)

	downloadPayload := encodeYtDLPPickerCallbackData(YtDLPActionDownload, sessionID, pickersession.YtDLPPickerTabNone, -1)
	backPayload := encodeYtDLPPickerCallbackData(YtDLPActionBack, sessionID, pickersession.YtDLPPickerTabNone, -1)

	rows = append(rows, markup.Row(markup.Data("Скачать", YtDLPPickerButtonUnique, downloadPayload)))
	rows = append(rows, markup.Row(markup.Data("Назад", YtDLPPickerButtonUnique, backPayload)))

	markup.Inline(rows...)

	message := fmt.Sprintf("Выбранный формат: %s \n Размер: %s \n Скачать?", option.DisplayName, formatFileSize(option.FileSize))

	return markup, message
}

// TODO: implement this
func (h *Handler) handleYoutubeMusicAndShortsRequest(c tele.Context, downloadCtx context.Context, statusMsg *tele.Message, user tele.Recipient, userID int64, url string, meta *ytdlp.Metadata) error {
	return nil
}

func encodeYtDLPPickerCallbackData(action, sessionID string, tab pickersession.YtDLPPickerTab, optionIdx int) string {
	if optionIdx >= 0 {
		return action + ":" + sessionID + ":" + string(tab) + ":" + strconv.Itoa(optionIdx)
	}
	if tab != pickersession.YtDLPPickerTabNone && tab != "" {
		return action + ":" + sessionID + ":" + string(tab)
	}
	return action + ":" + sessionID
}

func getYtDLPTabLabel(tab pickersession.YtDLPPickerTab) string {
	switch tab {
	case pickersession.YtDLPPickerTabAudioOnly:
		return "Аудио (только звук)"
	case pickersession.YtDLPPickerTabVideoOnly:
		return "Видео (только картинка)"
	case pickersession.YtDLPPickerTabAudioVideo:
		return "Аудио + Видео (звук и картинка в одном файле)"
	default:
		return "Неизвестно"
	}
}
