package handlers

import (
	"context"

	tele "gopkg.in/telebot.v4"
)

func (h *Handler) runDownloadJob(userID int64, fn func(context.Context) error) error {
	return h.userJobGuard.Run(userID, func() error {
		downloadCtx, cancel := context.WithTimeout(h.appCtx, h.downloadTimeout)
		defer cancel()

		return fn(downloadCtx)
	})
}

func editMessageText(c tele.Context, statusMsg *tele.Message, errMsg string) error {
	_, err := c.Bot().Edit(statusMsg, errMsg)
	return err
}
