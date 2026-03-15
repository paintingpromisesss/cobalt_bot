package handlers

import (
	"context"
	"time"

	"github.com/paintingpromisesss/cobalt_bot/internal/cobalt"
	"github.com/paintingpromisesss/cobalt_bot/internal/downloader"
	"github.com/paintingpromisesss/cobalt_bot/internal/queue"
	"github.com/paintingpromisesss/cobalt_bot/internal/storage"
	"github.com/paintingpromisesss/cobalt_bot/internal/telegram"
	pickersession "github.com/paintingpromisesss/cobalt_bot/internal/telegram/picker_session"
	"github.com/paintingpromisesss/cobalt_bot/internal/telegram/sender"
	"github.com/paintingpromisesss/cobalt_bot/internal/urlvalidator"
	"github.com/paintingpromisesss/cobalt_bot/internal/ytdlp"
	"go.uber.org/zap"
	tele "gopkg.in/telebot.v4"
)

type Handler struct {
	appCtx               context.Context
	requestTimeout       time.Duration
	downloadTimeout      time.Duration
	tb                   *telegram.Bot
	storage              *storage.DB
	queueManager         *queue.RequestQueue
	logger               *zap.Logger
	cobaltClient         *cobalt.CobaltClient
	downloader           *downloader.Downloader
	ytDLPClient          *ytdlp.Client
	urlValidator         *urlvalidator.URLValidator
	sender               *sender.FileSender
	availableServices    []string
	pickerSessionManager *pickersession.PickerSessionManager
}

func NewHandler(appCtx context.Context, requestTimeout time.Duration, downloadTimeout time.Duration, tb *telegram.Bot, storage *storage.DB, queueManager *queue.RequestQueue, logger *zap.Logger, cobaltClient *cobalt.CobaltClient, downloader *downloader.Downloader, ytDLPClient *ytdlp.Client, urlValidator *urlvalidator.URLValidator, sender *sender.FileSender, availableServices []string, pickerSessionManager *pickersession.PickerSessionManager) *Handler {
	return &Handler{
		appCtx:               appCtx,
		requestTimeout:       requestTimeout,
		downloadTimeout:      downloadTimeout,
		tb:                   tb,
		storage:              storage,
		queueManager:         queueManager,
		logger:               logger,
		cobaltClient:         cobaltClient,
		downloader:           downloader,
		ytDLPClient:          ytDLPClient,
		urlValidator:         urlValidator,
		sender:               sender,
		availableServices:    availableServices,
		pickerSessionManager: pickerSessionManager,
	}
}

func (h *Handler) RegisterHandlers() error {
	h.tb.Bot.Handle("/start", h.handleStart)
	h.tb.Bot.Handle(tele.OnText, h.handleMessage)
	h.tb.Bot.Handle(&tele.Btn{Unique: CobaltPickerButtonUnique}, h.handleCobaltPickerCallback)
	h.tb.Bot.Handle(&tele.Btn{Unique: YtDLPPickerButtonUnique}, h.handleYtDLPPickerCallback)
	return nil
}
