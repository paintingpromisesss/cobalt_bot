package app

import (
	"context"
	"errors"
	"os/signal"
	"syscall"

	"github.com/paintingpromisesss/cobalt_bot/internal/cobalt"
	"github.com/paintingpromisesss/cobalt_bot/internal/config"
	"github.com/paintingpromisesss/cobalt_bot/internal/downloader"
	"github.com/paintingpromisesss/cobalt_bot/internal/logger"
	"github.com/paintingpromisesss/cobalt_bot/internal/queue"
	"github.com/paintingpromisesss/cobalt_bot/internal/storage"
	"github.com/paintingpromisesss/cobalt_bot/internal/telegram"
	"github.com/paintingpromisesss/cobalt_bot/internal/telegram/handlers"
	pickersession "github.com/paintingpromisesss/cobalt_bot/internal/telegram/picker_session"
	"github.com/paintingpromisesss/cobalt_bot/internal/telegram/sender"
	"github.com/paintingpromisesss/cobalt_bot/internal/urlvalidator"
	"github.com/paintingpromisesss/cobalt_bot/internal/ytdlp"
	"go.uber.org/zap"
)

func Run(cfg config.Config) error {
	log, err := logger.New(cfg.Logging.Level)
	if err != nil {
		return err
	}
	defer func() {
		if syncErr := log.Sync(); syncErr != nil &&
			!errors.Is(syncErr, syscall.ENOTTY) &&
			!errors.Is(syncErr, syscall.EINVAL) {
			log.Warn("logger sync failed", zap.Error(syncErr))
		}
	}()

	storage, err := storage.New(cfg.Storage.DBPath)
	if err != nil {
		log.Error("init db failed", zap.Error(err))
		return err
	}
	defer func() {
		if closeErr := storage.Close(); closeErr != nil {
			log.Warn("db close failed", zap.Error(closeErr))
		}
	}()

	log.Info(
		"config loaded",
		zap.String("cobalt_base_url", cfg.Cobalt.BaseURL),
		zap.Int64("max_file_bytes", cfg.Storage.MaxFileBytes),
		zap.String("db_path", cfg.Storage.DBPath),
		zap.String("temp_dir", cfg.Storage.TempDir),
		zap.Duration("request_timeout", cfg.Timeouts.Request),
		zap.Duration("download_timeout", cfg.Timeouts.Download),
		zap.Duration("telegram_send_timeout", cfg.Timeouts.TelegramSend),
		zap.Duration("ffprobe_timeout", cfg.Timeouts.FFprobe),
		zap.Duration("ffmpeg_timeout", cfg.Timeouts.FFmpeg),
		zap.String("log_level", cfg.Logging.Level),
	)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	tgBot, err := telegram.New(cfg.Telegram.BotToken, cfg.Telegram.BotAPIURL, cfg.Timeouts.TelegramSend, log)
	if err != nil {
		log.Error("init telegram bot failed", zap.Error(err))
		return err
	}

	queueManager := queue.NewRequestQueue()

	cobaltClient := cobalt.NewCobaltClient(cfg.Cobalt.BaseURL, cfg.Timeouts.Request)

	downloader := downloader.NewDownloader(cfg.Timeouts.Download, cfg.Storage.TempDir, cfg.Storage.MaxFileBytes)
	ytDLPClient := ytdlp.NewClient(cfg.Storage.TempDir, cfg.YTDLP.MaxMediaDurationSeconds, cfg.Storage.MaxFileBytes, cfg.YTDLP.CurrentlyLiveAvailable, cfg.YTDLP.PlaylistAvailable)
	sender := sender.NewFileSender(log, cfg.Timeouts.FFprobe, cfg.Timeouts.FFmpeg)

	instanceInfo, err := cobaltClient.GetInstanceInfo(ctx)
	if err != nil {
		log.Error("get instance info failed", zap.Error(err))
		return err
	}

	availableServices := instanceInfo.Cobalt.Services
	urlValidator := urlvalidator.NewURLValidator(availableServices)
	pickerSessionManager := pickersession.NewPickerSessionManager(ctx, cfg.PickerSession.TTL, cfg.PickerSession.CleanupInterval)

	handler := handlers.NewHandler(
		ctx,
		cfg.Timeouts.Request,
		cfg.Timeouts.Download,
		cfg.YTDLP.MaxMediaDurationSeconds,
		tgBot,
		storage,
		queueManager,
		log,
		cobaltClient,
		downloader,
		ytDLPClient,
		urlValidator,
		sender,
		availableServices,
		pickerSessionManager,
	)
	if err := handler.RegisterHandlers(); err != nil {
		log.Error("register handlers failed", zap.Error(err))
		return err
	}

	tgBot.Run(ctx)
	log.Info("shutdown signal received")
	return nil
}
