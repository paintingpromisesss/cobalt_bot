package handlers

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/paintingpromisesss/cobalt_bot/internal/downloader"
	"github.com/paintingpromisesss/cobalt_bot/internal/telegram"
	pickersession "github.com/paintingpromisesss/cobalt_bot/internal/telegram/picker_session"
	"github.com/paintingpromisesss/cobalt_bot/internal/ytdlp"
	"go.uber.org/zap"
	tele "gopkg.in/telebot.v4"
)

func formatAvailableServices(services []string) string {
	result := ""
	for i, service := range services {
		result += service
		if i != len(services)-1 {
			result += ", "
		}
	}
	return result
}

func cleanupTempFile(log *zap.Logger, filePath string) {
	if filePath == "" {
		return
	}
	if err := os.Remove(filePath); err != nil && !errors.Is(err, os.ErrNotExist) {
		log.Warn("failed to remove temp file", zap.String("path", filePath), zap.Error(err))
	}
}

func formatFileSize(size int64) string {
	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
	)

	switch {
	case size >= GB:
		return formatFloatTrimmed(float64(size)/GB) + " GB"
	case size >= MB:
		return formatFloatTrimmed(float64(size)/MB) + " MB"
	case size >= KB:
		return formatFloatTrimmed(float64(size)/KB) + " KB"
	default:
		return strconv.FormatInt(size, 10) + " B"
	}
}

func formatFloatTrimmed(f float64) string {
	s := strconv.FormatFloat(f, 'f', 2, 64)
	s = strings.TrimRight(s, "0")
	s = strings.TrimRight(s, ".")
	return s
}

func formatDuration(duration int) string {
	totalSeconds := duration
	hours := totalSeconds / 3600
	minutes := (totalSeconds % 3600) / 60
	seconds := totalSeconds % 60

	if hours > 0 {
		return fmt.Sprintf("%d:%02d:%02d", hours, minutes, seconds)
	}
	return fmt.Sprintf("%d:%02d", minutes, seconds)
}

func formatDurationLimit(seconds int) string {
	if seconds <= 0 {
		return formatDuration(0)
	}
	return formatDuration(seconds)
}

func (h *Handler) handlePickerError(c tele.Context, statusMsg *tele.Message, err error) error {
	switch {
	case errors.Is(err, pickersession.ErrSessionExpired):
		_, err := c.Bot().Edit(statusMsg, "Время сессии истекло. Пожалуйста, попробуйте отправить ссылку заново.")
		return err
	case errors.Is(err, pickersession.ErrNoOptionsSelected):
		_, err := c.Bot().Edit(statusMsg, "Вы не выбрали ни одного объекта для загрузки. Пожалуйста, выберите хотя бы один и попробуйте снова.")
		return err
	default:
		_, err := c.Bot().Edit(statusMsg, h.pickerErrorToText(err))
		return err
	}
}

func (h *Handler) handlePickerCallbackError(c tele.Context, statusMsg *tele.Message, err error) error {
	if editErr := h.handlePickerError(c, statusMsg, err); editErr != nil {
		return editErr
	}

	return telegram.MarkHandled(err)
}

func (h *Handler) pickerErrorToText(err error) string {
	errorText := "Произошла ошибка при обработке вашего запроса: " + err.Error()
	switch {
	case errors.Is(err, context.DeadlineExceeded):
		errorText = "Не удалось завершить обработку вовремя. Попробуйте еще раз."
	case errors.Is(err, downloader.ErrFileTooLarge):
		errorText = "Файл слишком большой для отправки."
	case errors.Is(err, downloader.ErrEmptyFile):
		errorText = "Скачанный файл оказался пустым. Попробуйте повторить позже."
	case errors.Is(err, ytdlp.ErrMediaDurationTooLong):
		errorText = "Продолжительность медиафайла превышает допустимый лимит: " + formatDurationLimit(h.maxMediaDurationSecs) + "."
	}

	return errorText
}
