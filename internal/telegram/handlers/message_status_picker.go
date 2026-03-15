package handlers

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/paintingpromisesss/cobalt_bot/internal/cobalt"
	"github.com/paintingpromisesss/cobalt_bot/internal/downloader"
	pickersession "github.com/paintingpromisesss/cobalt_bot/internal/telegram/picker_session"
	tele "gopkg.in/telebot.v4"
)

const (
	ToggleAction    = "toggle"
	SelectAllAction = "select_all"
	ClearAllAction  = "clear_all"
	DownloadAction  = "download"
	CancelAction    = "cancel"

	maxAlbumFiles = 10
)

// handleMessageStatusPicker реализует обработку статуса Picker от Cobalt, который возвращает список объектов для скачивания.
func (h *Handler) handleMessageStatusPicker(c tele.Context, statusMsg *tele.Message, userID int64, cobaltResponse cobalt.MainResponse) error {
	pickerSessionID, err := h.pickerSessionManager.CreateCobaltSession(userID, cobaltResponse)
	if err != nil {
		return fmt.Errorf("create cobalt picker session: %w", err)
	}
	pickerView, err := h.pickerSessionManager.GetCobaltPickerView(pickerSessionID, userID)
	if err != nil {
		return err
	}

	return h.renderPickerKeyboard(c, statusMsg, pickerSessionID, &pickerView)

}

func (h *Handler) renderPickerKeyboard(c tele.Context, statusMsg *tele.Message, sessionID string, pickerView *pickersession.CobaltPickerView) error {
	markup, message := buildPickerMessage(sessionID, pickerView)
	_, err := c.Bot().Edit(statusMsg, message, &tele.SendOptions{ReplyMarkup: markup})
	return err
}

func (h *Handler) DownloadAndSendSelectedOptions(c tele.Context, statusMsg *tele.Message, downloadCtx context.Context, userID int64, user tele.Recipient, options []pickersession.CobaltPickerOption) error {
	if len(options) == 0 {
		return pickersession.ErrNoOptionsSelected
	}

	if _, err := c.Bot().Edit(statusMsg, fmt.Sprintf("Выбрано файлов: %d. Начинаю загрузку...", len(options))); err != nil {
		return err
	}

	downloadResults := make([]downloader.DownloadResult, 0, len(options))
	for _, option := range options {
		result, err := h.downloader.Download(downloadCtx, option.URL, option.Filename)
		if err != nil {
			for _, obj := range downloadResults {
				cleanupTempFile(h.logger, obj.Path)
			}
			return err
		}
		downloadResults = append(downloadResults, result)
	}
	defer func() {
		for _, obj := range downloadResults {
			cleanupTempFile(h.logger, obj.Path)
		}
	}()

	for _, result := range downloadResults {
		if result.Size <= 0 {
			return fmt.Errorf("downloaded file is empty: %s", result.Filename)
		}
	}

	if len(downloadResults) == 1 {
		result := downloadResults[0]
		if _, err := c.Bot().Edit(statusMsg, "Загрузка завершена. Отправляю файл..."); err != nil {
			return err
		}
		if err := h.sender.SendFile(c, result.Path, result.Filename, result.DetectedMIME, user); err != nil {
			return err
		}
		return nil
	}

	if _, err := c.Bot().Edit(statusMsg, "Загрузка завершена. Отправляю файлы..."); err != nil {
		return err
	}

	for start := 0; start < len(downloadResults); start += maxAlbumFiles {
		end := start + maxAlbumFiles
		if end > len(downloadResults) {
			end = len(downloadResults)
		}

		album := make(tele.Album, 0, end-start)
		for _, result := range downloadResults[start:end] {
			album = append(album, buildAlbumItem(result.Path, result.Filename, result.DetectedMIME))
		}

		if _, err := c.Bot().SendAlbum(user, album); err != nil {
			return fmt.Errorf("failed to send album: %w", err)
		}
	}

	if _, err := c.Bot().Edit(statusMsg, fmt.Sprintf("Готово. Отправлено файлов: %d.", len(downloadResults))); err != nil {
		return err
	}
	return nil
}

func buildPickerMessage(sessionID string, pickerView *pickersession.CobaltPickerView) (*tele.ReplyMarkup, string) {
	markup := &tele.ReplyMarkup{}
	total := len(pickerView.Options)
	rows := make([]tele.Row, 0, total+3)
	selected := 0

	for i, option := range pickerView.Options {
		indicator := NeutralIndicator
		if option.Selected {
			indicator = SelectedIndicator
			selected++
		}
		payload := encodeCobaltPickerCallbackData(ToggleAction, sessionID, i)
		rows = append(rows, markup.Row(markup.Data(indicator+" "+option.Label, CobaltPickerButtonUnique, payload)))

	}

	if selected > 0 {
		rows = append(rows, markup.Row(
			markup.Data(UnselectedIndicator+" Очистить выбор", CobaltPickerButtonUnique, encodeCobaltPickerCallbackData(ClearAllAction, sessionID, -1)),
			markup.Data(DownloadIndicator+" Скачать", CobaltPickerButtonUnique, encodeCobaltPickerCallbackData(DownloadAction, sessionID, -1)),
		))
	} else {
		rows = append(rows, markup.Row(markup.Data(SelectedIndicator+" Выбрать все", CobaltPickerButtonUnique, encodeCobaltPickerCallbackData(SelectAllAction, sessionID, -1))))
	}

	rows = append(rows, markup.Row(markup.Data(UnselectedIndicator+" Отменить", CobaltPickerButtonUnique, encodeCobaltPickerCallbackData(CancelAction, sessionID, -1))))

	markup.Inline(rows...)

	message := fmt.Sprintf("Найдено файлов: %d. Выбрано: %d.\n Отметьте нужные и нажмите «Скачать».", total, selected)
	return markup, message
}

func buildAlbumItem(filepath, filename, detectedMIME string) tele.Inputtable {
	file := tele.FromDisk(filepath)
	mime := strings.TrimSpace(strings.ToLower(detectedMIME))

	switch {
	case strings.HasPrefix(mime, "image/"):
		return &tele.Photo{File: file}
	case strings.HasPrefix(mime, "video/"):
		return &tele.Video{File: file, FileName: filename, MIME: detectedMIME, Streaming: true}
	default:
		return &tele.Document{File: file, FileName: filename, MIME: detectedMIME}
	}
}

func encodeCobaltPickerCallbackData(action, sessionID string, optionIdx int) string {
	if optionIdx >= 0 {
		return action + ":" + sessionID + ":" + strconv.Itoa(optionIdx)
	}
	return action + ":" + sessionID
}
