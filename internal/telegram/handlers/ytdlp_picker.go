package handlers

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/paintingpromisesss/cobalt_bot/internal/telegram"
	pickersession "github.com/paintingpromisesss/cobalt_bot/internal/telegram/picker_session"
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
			return handlePickerCallbackError(c, statusMsg, err)
		}
		return h.renderYtDLPPickerKeyboard(c, statusMsg, sessionID, &pickerView)
	case YtDLPActionChoose:
		pickerOption, err := h.pickerSessionManager.ChooseYtDLPOption(sessionID, userID, optionIdx)
		if err != nil {
			return handlePickerCallbackError(c, statusMsg, err)
		}
		return h.renderYtDLPConfirmationKeyboard(c, statusMsg, sessionID, pickerOption)
	case YtDLPActionDownload:
		pickerOption, err := h.pickerSessionManager.ConsumeChosenYtDLPOption(sessionID, userID)
		if err != nil {
			return handlePickerCallbackError(c, statusMsg, err)
		}

		err = h.queueManager.Run(userID, func() error {
			downloadCtx, cancel := context.WithTimeout(h.appCtx, h.downloadTimeout)
			defer cancel()

			return h.DownloadAndSendYtDLPOption(c, downloadCtx, statusMsg, userID, c.Recipient(), pickerOption)
		})
		if err != nil {
			return handlePickerCallbackError(c, statusMsg, err)
		}
		return nil
	case YtDLPActionBack:
		pickerView, err := h.pickerSessionManager.ClearChosenYtDLPOption(sessionID, userID)
		if err != nil {
			return handlePickerCallbackError(c, statusMsg, err)
		}
		return h.renderYtDLPPickerKeyboard(c, statusMsg, sessionID, &pickerView)
	case YtDLPActionCancel:
		err := h.pickerSessionManager.DeleteSession(sessionID, userID, pickersession.PickerSessionTypeYtDLP)
		if err != nil {
			return handlePickerCallbackError(c, statusMsg, err)
		}
		_, err = c.Bot().Edit(statusMsg, "Сессия выбора отменена. Если хотите скачать что-то ещё, просто отправьте ссылку.")
		return err
	default:
		_, err := c.Bot().Edit(statusMsg, "Неизвестное действие. Пожалуйста, попробуйте снова.")
		return err
	}
}

func (h *Handler) DownloadAndSendYtDLPOption(c tele.Context, downloadCtx context.Context, statusMsg *tele.Message, userID int64, user tele.Recipient, option pickersession.YtDLPPickerOption) error {
	if _, err := c.Bot().Edit(statusMsg, fmt.Sprintf("Начинаю загрузку формата: %s...", option.DisplayName)); err != nil {
		return err
	}

	urls := option.GetURLsToDownload()
	urlsLen := urls.GetLen()
	if urlsLen == 0 {
		return fmt.Errorf("no URLs to download for the selected option: %s", option.DisplayName)
	}

	if urlsLen > 2 {
		return fmt.Errorf("too many URLs to download for the selected option: %s", option.DisplayName)
	}

	if urlsLen == 1 {
		url := urls.GetSingleURL()
		if url == "" {
			return fmt.Errorf("no valid URL found for the selected option: %s", option.DisplayName)
		}

		result, err := h.downloader.Download(downloadCtx, url, option.DisplayName)
		if err != nil {
			return err
		}
		defer cleanupTempFile(h.logger, result.Path)

		if result.Size <= 0 {
			return fmt.Errorf("downloaded file is empty: %s", result.Filename)
		}
		if _, err := c.Bot().Edit(statusMsg, "Загрузка завершена. Отправляю файл..."); err != nil {
			return err
		}
		if err := h.sender.SendFile(c, result.Path, result.Filename, result.DetectedMIME, user); err != nil {
			return err
		}
		return nil
	}

	if urlsLen == 2 {
		audioURL := urls.AudioURL
		videoURL := urls.VideoURL

		if audioURL == nil || videoURL == nil {
			return fmt.Errorf("both audio and video URLs must be present for muxed download")
		}

		audioResult, err := h.downloader.Download(downloadCtx, *audioURL, option.DisplayName+"_audio")
		if err != nil {
			return err
		}
		defer cleanupTempFile(h.logger, audioResult.Path)

		videoResult, err := h.downloader.Download(downloadCtx, *videoURL, option.DisplayName+"_video")
		if err != nil {
			return err
		}
		defer cleanupTempFile(h.logger, videoResult.Path)

		if audioResult.Size <= 0 {
			return fmt.Errorf("downloaded audio file is empty: %s", audioResult.Filename)
		}
		if videoResult.Size <= 0 {
			return fmt.Errorf("downloaded video file is empty: %s", videoResult.Filename)
		}

		if _, err := c.Bot().Edit(statusMsg, "Загрузка завершена. Обрабатываю файлы..."); err != nil {
			return err
		}

		mergedMP4Path, mergedMetadata, err := h.sender.MergeStreamsToStreamableMP4(videoResult.Path, audioResult.Path)
		if err != nil {
			return fmt.Errorf("failed to merge audio and video streams: %w", err)
		}
		defer cleanupTempFile(h.logger, mergedMP4Path)

		if mergedMetadata.Size <= 0 {
			return fmt.Errorf("merged file is empty: %s", mergedMP4Path)
		}

		if _, err := c.Bot().Edit(statusMsg, fmt.Sprintf("Обработка завершена. Размер: %s, длительность: %s, разрешение: %dx%d (%s, %s). Отправляю файл...", formatFileSize(mergedMetadata.Size), formatDuration(mergedMetadata.Duration), mergedMetadata.Width, mergedMetadata.Height, mergedMetadata.VideoCodec, mergedMetadata.AudioCodec)); err != nil {
			return err
		}
		if err := h.sender.SendFile(c, mergedMP4Path, option.DisplayName+".mp4", mergedMetadata.MIME, user); err != nil {
			return err
		}
		return nil
	}
	return nil
}

func parseYtDLPPickerCallbackData(data string) (action, sessionID string, tab pickersession.YtDLPPickerTab, optionIdx int, err error) {
	parts := strings.Split(strings.TrimSpace(data), ":")
	if len(parts) < 2 || len(parts) > 4 {
		return "", "", pickersession.YtDLPPickerTabNone, -1, fmt.Errorf("invalid callback data format")
	}

	action, sessionID, tab, optionIdx = parts[0], parts[1], pickersession.YtDLPPickerTabNone, -1
	if len(parts) == 4 {
		tab = pickersession.YtDLPPickerTab(parts[2])
		idx, convErr := strconv.Atoi(parts[3])
		if convErr != nil {
			return "", "", pickersession.YtDLPPickerTabNone, -1, fmt.Errorf("invalid option index: %v", convErr)
		}
		optionIdx = idx
	}
	return action, sessionID, tab, optionIdx, nil
}
