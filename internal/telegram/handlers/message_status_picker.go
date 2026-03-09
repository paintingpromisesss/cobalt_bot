package handlers

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/paintingpromisesss/cobalt_bot/internal/cobalt"
	"github.com/paintingpromisesss/cobalt_bot/internal/downloader"
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
	CancelAction    = "cancel"

	PickerButtonUnique = "picker_button"
	maxAlbumFiles      = 10
)

// handleMessageStatusPicker реализует обработку статуса Picker от Cobalt, который возвращает список объектов для скачивания.
func (h *Handler) handleMessageStatusPicker(c tele.Context, statusMsg *tele.Message, userID int64, cobaltResponse cobalt.MainResponse) error {
	pickerSessionID := h.pickerSessionManager.CreateSession(userID, cobaltResponse)
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

func (h *Handler) DownloadSelectedOptions(c tele.Context, statusMsg *tele.Message, downloadCtx context.Context, userID int64, options []pickersession.PickerOption) error {
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
		if err := h.sender.SendFile(c, result.Path, result.Filename, result.DetectedMIME, statusMsg); err != nil {
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

		if _, err := c.Bot().SendAlbum(c.Recipient(), album); err != nil {
			return fmt.Errorf("failed to send album: %w", err)
		}
	}

	if _, err := c.Bot().Edit(statusMsg, fmt.Sprintf("Готово. Отправлено файлов: %d.", len(downloadResults))); err != nil {
		return err
	}
	return nil
}

func buildPickerMessage(sessionID string, pickerView *pickersession.PickerView) (*tele.ReplyMarkup, string) {
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

	rows = append(rows, markup.Row(markup.Data(UnselectedIndicator+" Отменить", PickerButtonUnique, encodePickerCallbackData(CancelAction, sessionID, -1))))

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

func encodePickerCallbackData(action, sessionID string, optionIdx int) string {
	if optionIdx >= 0 {
		return action + ":" + sessionID + ":" + strconv.Itoa(optionIdx)
	}
	return action + ":" + sessionID
}

func parsePickerCallbackData(data string) (action, sessionID string, optionIdx int, err error) {
	parts := strings.Split(strings.TrimSpace(data), ":")
	if len(parts) < 2 || len(parts) > 3 {
		return "", "", 0, fmt.Errorf("invalid callback data format")
	}

	action, sessionID, optionIdx = parts[0], parts[1], -1
	if len(parts) == 3 {
		idx, convErr := strconv.Atoi(parts[2])
		if convErr != nil {
			return "", "", 0, fmt.Errorf("invalid option index: %v", convErr)
		}
		optionIdx = idx
	}
	return action, sessionID, optionIdx, nil
}

func handlePickerError(c tele.Context, statusMsg *tele.Message, err error) error {
	switch {
	case errors.Is(err, pickersession.ErrSessionExpired):
		_, err := c.Bot().Edit(statusMsg, "Время сессии истекло. Пожалуйста, попробуйте отправить ссылку заново.")
		return err
	case errors.Is(err, pickersession.ErrNoOptionsSelected):
		_, err := c.Bot().Edit(statusMsg, "Вы не выбрали ни одного объекта для загрузки. Пожалуйста, выберите хотя бы один и попробуйте снова.")
		return err
	default:
		_, err := c.Bot().Edit(statusMsg, pickerErrorToText(err))
		return err
	}
}

func pickerErrorToText(err error) string {
	errorText := "Произошла ошибка при обработке вашего запроса: " + err.Error()
	switch {
	case errors.Is(err, context.DeadlineExceeded):
		errorText = "Не удалось завершить обработку вовремя. Попробуйте еще раз."
	case errors.Is(err, downloader.ErrFileTooLarge):
		errorText = "Файл слишком большой для отправки."
	case errors.Is(err, downloader.ErrEmptyFile):
		errorText = "Скачанный файл оказался пустым. Попробуйте повторить позже."
	}

	return errorText
}
