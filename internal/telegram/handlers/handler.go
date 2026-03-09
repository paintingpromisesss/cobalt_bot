package handlers

import (
	"context"

	"github.com/paintingpromisesss/cobalt_bot/internal/cobalt"
	"github.com/paintingpromisesss/cobalt_bot/internal/downloader"
	"github.com/paintingpromisesss/cobalt_bot/internal/queue"
	"github.com/paintingpromisesss/cobalt_bot/internal/storage"
	"github.com/paintingpromisesss/cobalt_bot/internal/telegram"
	pickersession "github.com/paintingpromisesss/cobalt_bot/internal/telegram/picker_session"
	"github.com/paintingpromisesss/cobalt_bot/internal/telegram/sender"
	"github.com/paintingpromisesss/cobalt_bot/internal/urlvalidator"
	"go.uber.org/zap"
	tele "gopkg.in/telebot.v4"
)

type Handler struct {
	appCtx               context.Context
	tb                   *telegram.Bot
	storage              *storage.DB
	queueManager         *queue.RequestQueue
	logger               *zap.Logger
	cobaltClient         *cobalt.CobaltClient
	downloader           *downloader.Downloader
	urlValidator         *urlvalidator.URLValidator
	sender               *sender.FileSender
	availableServices    []string
	pickerSessionManager *pickersession.PickerSessionManager
}

func NewHandler(appCtx context.Context, tb *telegram.Bot, storage *storage.DB, queueManager *queue.RequestQueue, logger *zap.Logger, cobaltClient *cobalt.CobaltClient, downloader *downloader.Downloader, urlValidator *urlvalidator.URLValidator, sender *sender.FileSender, availableServices []string, pickerSessionManager *pickersession.PickerSessionManager) *Handler {
	return &Handler{
		appCtx:               appCtx,
		tb:                   tb,
		storage:              storage,
		queueManager:         queueManager,
		logger:               logger,
		cobaltClient:         cobaltClient,
		downloader:           downloader,
		urlValidator:         urlValidator,
		sender:               sender,
		availableServices:    availableServices,
		pickerSessionManager: pickerSessionManager,
	}
}

func (h *Handler) RegisterHandlers() error {
	h.tb.Bot.Handle("/start", h.handleStart)
	h.tb.Bot.Handle(tele.OnText, h.handleMessage)
	h.tb.Bot.Handle(&tele.Btn{Unique: PickerButtonUnique}, h.handlePickerCallback)
	return nil
}
