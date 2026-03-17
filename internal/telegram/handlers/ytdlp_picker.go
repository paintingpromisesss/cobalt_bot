package handlers

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/paintingpromisesss/cobalt_bot/internal/telegram"
	pickersession "github.com/paintingpromisesss/cobalt_bot/internal/telegram/picker_session"
	"github.com/paintingpromisesss/cobalt_bot/internal/ytdlp"
	"go.uber.org/zap"
	tele "gopkg.in/telebot.v4"
)

func (h *Handler) handleYtDLPPickerCallback(c tele.Context) error {
	if err := c.Respond(); err != nil {
		h.logger.Warn("failed to respond to picker callback", zap.Error(err))
	}

	userID := c.Sender().ID
	statusMsg := c.Message()

	action, sessionID, tab, optionIdx, err := parseYtDLPPickerCallbackData(c.Data())
	if err != nil {
		h.logger.Warn("failed to parse picker callback data", zap.Int64("user_id", userID), zap.String("data", c.Data()), zap.Error(err))

		if err := c.Edit("Не удалось распознать действие. Пожалуйста, попробуйте снова."); err != nil {
			return err
		}

		return telegram.MarkHandled(err)
	}

	switch action {
	case YtDLPActionTab:
		pickerView, err := h.pickerSessionManager.SetYtDLPActiveTab(sessionID, userID, tab)
		if err != nil {
			return h.handlePickerCallbackError(c, statusMsg, err)
		}
		return h.renderYtDLPPickerKeyboard(c, statusMsg, sessionID, &pickerView)
	case YtDLPActionChoose:
		pickerOption, err := h.pickerSessionManager.ChooseYtDLPOption(sessionID, userID, optionIdx)
		if err != nil {
			return h.handlePickerCallbackError(c, statusMsg, err)
		}
		return h.renderYtDLPConfirmationKeyboard(c, statusMsg, sessionID, pickerOption)
	case YtDLPActionDownload:
		pickerOption, err := h.pickerSessionManager.ConsumeChosenYtDLPOption(sessionID, userID)
		if err != nil {
			return h.handlePickerCallbackError(c, statusMsg, err)
		}

		err = h.queueManager.Run(userID, func() error {
			downloadCtx, cancel := context.WithTimeout(h.appCtx, h.downloadTimeout)
			defer cancel()

			return h.DownloadAndSendYtDLPOption(c, downloadCtx, statusMsg, c.Recipient(), pickerOption)
		})
		if err != nil {
			return h.handlePickerCallbackError(c, statusMsg, err)
		}
		return nil
	case YtDLPActionConfirmBack:
		pickerView, err := h.pickerSessionManager.ClearChosenYtDLPOption(sessionID, userID)
		if err != nil {
			return h.handlePickerCallbackError(c, statusMsg, err)
		}
		return h.renderYtDLPPickerKeyboard(c, statusMsg, sessionID, &pickerView)
	case YtDLPActionBack:
		pickerView, err := h.pickerSessionManager.SetYtDLPActiveTab(sessionID, userID, tab)
		if err != nil {
			return h.handlePickerCallbackError(c, statusMsg, err)
		}
		return h.renderYtDLPPickerKeyboard(c, statusMsg, sessionID, &pickerView)
	case YtDLPActionCancel:
		err := h.pickerSessionManager.DeleteSession(sessionID, userID, pickersession.PickerSessionTypeYtDLP)
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

func (h *Handler) DownloadAndSendYtDLPOption(c tele.Context, downloadCtx context.Context, statusMsg *tele.Message, user tele.Recipient, option pickersession.YtDLPPickerOption) error {
	if _, err := c.Bot().Edit(statusMsg, fmt.Sprintf("Начинаю загрузку формата: %s...", option.DisplayName)); err != nil {
		return err
	}

	var selectedFormat *ytdlp.Format
	if option.Format.FormatID != "" {
		selectedFormat = &option.Format
	}

	downloadResult, err := h.ytDLPClient.Download(downloadCtx, option.ContentURL, option.FormatID, selectedFormat)
	if err != nil {
		return err
	}
	defer cleanupTempFile(h.logger, downloadResult.Path)

	if downloadResult.Size <= 0 {
		return fmt.Errorf("downloaded file is empty: %s", downloadResult.Filename)
	}

	if _, err := c.Bot().Edit(statusMsg, "Загрузка завершена. Отправляю файл..."); err != nil {
		return err
	}

	err = h.sender.SendFile(c, downloadResult.Path, downloadResult.Filename, downloadResult.DetectedMIME, user)
	if err != nil {
		return err
	}

	return nil
}

func parseYtDLPPickerCallbackData(data string) (action, sessionID string, tab pickersession.YtDLPPickerTab, optionIdx int, err error) {
	parts := strings.Split(strings.TrimSpace(data), ":")
	if len(parts) < 2 || len(parts) > 4 {
		return "", "", pickersession.YtDLPPickerTabNone, -1, fmt.Errorf("invalid callback data format")
	}

	action, sessionID, tab, optionIdx = parts[0], parts[1], pickersession.YtDLPPickerTabNone, -1
	if len(parts) >= 3 {
		tab = pickersession.YtDLPPickerTab(parts[2])
	}
	if len(parts) == 4 {
		idx, convErr := strconv.Atoi(parts[3])
		if convErr != nil {
			return "", "", pickersession.YtDLPPickerTabNone, -1, fmt.Errorf("invalid option index: %v", convErr)
		}
		optionIdx = idx
	}
	return action, sessionID, tab, optionIdx, nil
}
