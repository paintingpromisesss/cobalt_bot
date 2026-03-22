package handlers

import (
	"github.com/paintingpromisesss/cobalt_bot/internal/telegram"
	"go.uber.org/zap"
	tele "gopkg.in/telebot.v4"
)

func (h *Handler) respondToCallback(c tele.Context) {
	if err := c.Respond(); err != nil {
		h.logger.Warn("failed to respond to picker callback", zap.Error(err))
	}
}

func (h *Handler) handleInvalidCallbackData(c tele.Context, userID int64, data string, err error) error {
	h.logger.Warn("failed to parse picker callback data", zap.Int64("user_id", userID), zap.String("data", data), zap.Error(err))

	if editErr := c.Edit("Не удалось распознать действие. Пожалуйста, попробуйте снова."); editErr != nil {
		return editErr
	}

	return telegram.MarkHandled(err)
}
