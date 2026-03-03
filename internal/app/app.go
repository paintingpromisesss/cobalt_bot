package app

import (
	"context"
	"errors"
	"fmt"
	"os/signal"
	"syscall"

	"github.com/paintingpromisesss/cobalt_bot/internal/config"
	"github.com/paintingpromisesss/cobalt_bot/internal/logger"
	"github.com/paintingpromisesss/cobalt_bot/internal/storage"
	"github.com/paintingpromisesss/cobalt_bot/internal/telegram"
	"go.uber.org/zap"
)

func Run(cfg config.Config) error {
	log, err := logger.New(cfg.LogLevel)
	if err != nil {
		return fmt.Errorf("init logger: %w", err)
	}
	defer func() {
		if syncErr := log.Sync(); syncErr != nil &&
			!errors.Is(syncErr, syscall.ENOTTY) &&
			!errors.Is(syncErr, syscall.EINVAL) {
			log.Warn("logger sync failed", zap.Error(syncErr))
		}
	}()

	sqliteDB, err := storage.New(cfg.DBPath)
	if err != nil {
		return fmt.Errorf("init db: %w", err)
	}
	defer func() {
		if closeErr := sqliteDB.Close(); closeErr != nil {
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
		return fmt.Errorf("init telegram bot: %w", err)
	}

	tgBot.Run(ctx)
	log.Info("shutdown signal received")
	return nil
}
