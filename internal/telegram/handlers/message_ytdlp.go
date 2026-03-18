package handlers

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/paintingpromisesss/cobalt_bot/internal/domain/media"
	"github.com/paintingpromisesss/cobalt_bot/internal/domain/picker"
	"github.com/paintingpromisesss/cobalt_bot/internal/ytdlp"
	tele "gopkg.in/telebot.v4"
)

const (
	YtDLPActionTab         = "select_tab"
	YtDLPActionChoose      = "choose"
	YtDLPActionDownload    = "download"
	YtDLPActionCancel      = "cancel"
	YtDLPActionConfirmBack = "confirm_back"
	YtDLPActionBack        = "back"
)

func (h *Handler) handleYoutubeVideoRequest(c tele.Context, statusMsg *tele.Message, userID int64, meta *ytdlp.Metadata) error {
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

func (h *Handler) renderYtDLPPickerKeyboard(c tele.Context, statusMsg *tele.Message, sessionID string, pickerView *picker.YtDLPView) error {
	var markup *tele.ReplyMarkup
	var message string
	if pickerView.ActiveTab == picker.YtDLPTabNone {
		markup, message = buildYtDLPTabsMessage(sessionID, pickerView)
	} else {
		markup, message = buildYtDLPOptionsMessage(sessionID, pickerView)
	}

	_, err := c.Bot().Edit(statusMsg, message, &tele.SendOptions{ReplyMarkup: markup})
	return err
}

func buildYtDLPTabsMessage(sessionID string, pickerView *picker.YtDLPView) (*tele.ReplyMarkup, string) {
	markup := &tele.ReplyMarkup{}
	total := len(pickerView.Tabs)
	rows := make([]tele.Row, 0, total+1)
	for _, tab := range pickerView.Tabs {
		payload := encodeYtDLPPickerCallbackData(YtDLPActionTab, sessionID, tab, -1)
		text := getYtDLPTabLabel(tab)

		rows = append(rows, markup.Row(markup.Data(text, YtDLPPickerButtonUnique, payload)))
	}

	payload := encodeYtDLPPickerCallbackData(YtDLPActionCancel, sessionID, picker.YtDLPTabNone, -1)
	rows = append(rows, markup.Row(markup.Data("Отменить", YtDLPPickerButtonUnique, payload)))

	markup.Inline(rows...)

	message := fmt.Sprintf("Скачиваемый контент: %s \n Выберите опцию скачивания:", pickerView.ContentName)

	return markup, message
}

func buildYtDLPOptionsMessage(sessionID string, pickerView *picker.YtDLPView) (*tele.ReplyMarkup, string) {
	markup := &tele.ReplyMarkup{}
	rows := make([]tele.Row, 0, len(pickerView.Options)+1)

	for i, option := range pickerView.Options {
		rows = append(rows, markup.Row(markup.Data(option.DisplayName, YtDLPPickerButtonUnique, encodeYtDLPPickerCallbackData(YtDLPActionChoose, sessionID, pickerView.ActiveTab, i))))
	}

	rows = append(rows, markup.Row(markup.Data("Назад", YtDLPPickerButtonUnique, encodeYtDLPPickerCallbackData(YtDLPActionBack, sessionID, picker.YtDLPTabNone, -1))))

	markup.Inline(rows...)

	message := fmt.Sprintf("Выберите формат скачивания для: %s\n (тип: %s)", pickerView.ContentName, pickerView.ActiveTab)

	return markup, message
}

func (h *Handler) renderYtDLPConfirmationKeyboard(c tele.Context, statusMsg *tele.Message, sessionID string, option picker.YtDLPOption) error {
	markup, message := buildYtDLPConfirmationMessage(sessionID, option)
	_, err := c.Bot().Edit(statusMsg, message, &tele.SendOptions{ReplyMarkup: markup})
	return err
}

func buildYtDLPConfirmationMessage(sessionID string, option picker.YtDLPOption) (*tele.ReplyMarkup, string) {
	markup := &tele.ReplyMarkup{}
	rows := make([]tele.Row, 0, 2)

	downloadPayload := encodeYtDLPPickerCallbackData(YtDLPActionDownload, sessionID, picker.YtDLPTabNone, -1)
	backPayload := encodeYtDLPPickerCallbackData(YtDLPActionConfirmBack, sessionID, picker.YtDLPTabNone, -1)

	rows = append(rows, markup.Row(markup.Data("Скачать", YtDLPPickerButtonUnique, downloadPayload)))
	rows = append(rows, markup.Row(markup.Data("Назад", YtDLPPickerButtonUnique, backPayload)))

	markup.Inline(rows...)

	message := fmt.Sprintf("Выбранный формат: %s\n Размер: %s\n Скачать?", option.DisplayName, formatFileSize(option.FileSize))

	return markup, message
}

func (h *Handler) handleYoutubeMusicRequest(c tele.Context, downloadCtx context.Context, statusMsg *tele.Message, user tele.Recipient, meta *ytdlp.Metadata) error {
	var bestAudioFormat *ytdlp.Format
	for _, requestedDownload := range meta.RequestedDownloads {
		bestAudioFormat = requestedDownload.GetBestAudioFormat()
		break
	}
	if bestAudioFormat == nil {
		return fmt.Errorf("не найден подходящий аудио формат для скачивания")
	}

	option := picker.YtDLPOption{
		DisplayName:  bestAudioFormat.GetDisplayName(nil, nil),
		ThumbnailURL: meta.Thumbnail,
		ContentURL:   meta.OriginalURL,
		FormatID:     bestAudioFormat.FormatID,
		FileSize:     bestAudioFormat.FileSize,
		Duration:     time.Duration(meta.Duration) * time.Second,
		Format:       media.DownloadFormat{HasAudio: true},
	}

	return h.DownloadAndSendYtDLPOption(c, downloadCtx, statusMsg, user, option)
}

// TODO: implement this
func (h *Handler) handleYoutubeShortsRequest(c tele.Context, downloadCtx context.Context, statusMsg *tele.Message, user tele.Recipient, meta *ytdlp.Metadata) error {
	var bestVideoFormat *ytdlp.Format
	var bestAudioFormat *ytdlp.Format
	for _, requestedDownload := range meta.RequestedDownloads {
		bestVideoFormat = requestedDownload.GetBestVideoFormat()
		bestAudioFormat = requestedDownload.GetBestAudioFormat()
		break
	}
	if bestVideoFormat == nil {
		return fmt.Errorf("не найден подходящий видео формат для скачивания")
	}
	if bestAudioFormat == nil {
		return fmt.Errorf("не найден подходящий аудио формат для скачивания")
	}

	option := picker.YtDLPOption{
		DisplayName:  bestVideoFormat.GetDisplayName(bestAudioFormat, bestVideoFormat),
		ThumbnailURL: meta.Thumbnail,
		ContentURL:   meta.OriginalURL,
		FormatID:     bestVideoFormat.FormatID + "+" + bestAudioFormat.FormatID,
		FileSize:     bestVideoFormat.FileSize + bestAudioFormat.FileSize,
		Duration:     time.Duration(meta.Duration) * time.Second,
		Format:       media.DownloadFormat{HasAudio: true, HasVideo: true},
	}
	return h.DownloadAndSendYtDLPOption(c, downloadCtx, statusMsg, user, option)
}

func encodeYtDLPPickerCallbackData(action, sessionID string, tab picker.YtDLPTab, optionIdx int) string {
	if optionIdx >= 0 {
		return action + ":" + sessionID + ":" + string(tab) + ":" + strconv.Itoa(optionIdx)
	}
	if tab != picker.YtDLPTabNone && tab != "" {
		return action + ":" + sessionID + ":" + string(tab)
	}
	return action + ":" + sessionID
}

func getYtDLPTabLabel(tab picker.YtDLPTab) string {
	switch tab {
	case picker.YtDLPTabAudioOnly:
		return "Аудио (только звук)"
	case picker.YtDLPTabVideoOnly:
		return "Видео (только картинка)"
	case picker.YtDLPTabAudioVideo:
		return "Аудио + Видео (звук и картинка в одном файле)"
	default:
		return "Неизвестно"
	}
}
