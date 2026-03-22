package handlers

import (
	"context"

	usecasestart "github.com/paintingpromisesss/cobalt_bot/internal/usecase/start"
	"go.uber.org/zap"
	tele "gopkg.in/telebot.v4"
)

func (h *Handler) handleStart(c tele.Context) error {
	reqCtx, cancel := context.WithTimeout(h.appCtx, h.requestTimeout)
	defer cancel()
	userID := c.Sender().ID
	result, err := h.startService.Handle(reqCtx, usecasestart.Input{UserID: userID})
	if err != nil {
		return err
	}
	h.logger.Info("user started the bot", zap.Int64("user_id", userID))
	return c.Send(result.Message)
}
