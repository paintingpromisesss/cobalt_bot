package handlers

import (
	"context"

	"github.com/paintingpromisesss/cobalt_bot/internal/ytdlp"
	tele "gopkg.in/telebot.v4"
)

// TODO: implement this
func (h *Handler) handleYoutubeVideoRequest(c tele.Context, downloadCtx context.Context, statusMsg *tele.Message, user tele.Recipient, userID int64, url string, meta *ytdlp.Metadata) error {
	return nil
}

// TODO: implement this
func (h *Handler) handleYoutubeMusicAndShortsRequest(c tele.Context, downloadCtx context.Context, statusMsg *tele.Message, user tele.Recipient, userID int64, url string, meta *ytdlp.Metadata) error {
	return nil
}
