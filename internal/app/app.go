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
	"go.uber.org/zap"
)

func Run(cfg config.Config) error {
	log, err := logger.New(cfg.LogLevel)
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

	storage, err := storage.New(cfg.DBPath)
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
		zap.String("cobalt_base_url", cfg.CobaltBaseURL),
		zap.Int64("max_file_bytes", cfg.MaxFileBytes),
		zap.String("db_path", cfg.DBPath),
		zap.String("temp_dir", cfg.TempDir),
		zap.Duration("request_timeout", cfg.RequestTimeout),
		zap.Duration("download_timeout", cfg.DownloadTimeout),
		zap.String("log_level", cfg.LogLevel),
	)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	tgBot, err := telegram.New(cfg.TelegramBotToken, log)
	if err != nil {
		log.Error("init telegram bot failed", zap.Error(err))
		return err
	}

	queueManager := queue.NewRequestQueue()

	cobaltClient := cobalt.NewCobaltClient(cfg.CobaltBaseURL, cfg.RequestTimeout)

	downloader := downloader.NewDownloader(cfg.DownloadTimeout, cfg.TempDir, cfg.MaxFileBytes)
	sender := sender.NewFileSender()

	instanceInfo, err := cobaltClient.GetInstanceInfo(ctx)
	if err != nil {
		log.Error("get instance info failed", zap.Error(err))
		return err
	}

	availableServices := instanceInfo.Cobalt.Services
	urlValidator := urlvalidator.NewURLValidator(availableServices)
	pickerSessionManager := pickersession.NewPickerSessionManager(cfg.PickerSessionManagerTTL)

	handler := handlers.NewHandler(ctx, tgBot, storage, queueManager, log, cobaltClient, downloader, urlValidator, sender, availableServices, pickerSessionManager)
	if err := handler.RegisterHandlers(); err != nil {
		log.Error("register handlers failed", zap.Error(err))
		return err
	}

	tgBot.Run(ctx)
	log.Info("shutdown signal received")
	return nil
}
