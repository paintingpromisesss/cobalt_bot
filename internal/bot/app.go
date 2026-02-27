package bot

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"

	"github.com/paintingpromisesss/cobalt_bot/internal/cobalt"
	"github.com/paintingpromisesss/cobalt_bot/internal/config"
	"github.com/paintingpromisesss/cobalt_bot/internal/download"
	"github.com/paintingpromisesss/cobalt_bot/internal/logger"
	"github.com/paintingpromisesss/cobalt_bot/internal/queue"
	"github.com/paintingpromisesss/cobalt_bot/internal/storage"
	"github.com/paintingpromisesss/cobalt_bot/internal/user_settings"
	"go.uber.org/zap"
)

func Run(cfg config.Config) error {
	log, err := logger.New(cfg.LogLevel)
	if err != nil {
		return fmt.Errorf("init logger: %w", err)
	}
	defer func() { _ = log.Sync() }()

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
		zap.String("mihomo_base_url", cfg.MihomoBaseURL),
		zap.Int64("max_file_bytes", cfg.MaxFileBytes),
		zap.String("db_path", cfg.DBPath),
		zap.String("temp_dir", cfg.TempDir),
		zap.Duration("request_timeout", cfg.RequestTimeout),
		zap.Duration("download_timeout", cfg.DownloadTimeout),
		zap.String("log_level", cfg.LogLevel),
	)

	cobaltClient, err := cobalt.NewClient(cfg.CobaltBaseURL, cfg.RequestTimeout)
	if err != nil {
		return fmt.Errorf("init cobalt client: %w", err)
	}
	_ = cobaltClient

	settingsRepo, err := storage.NewUserSettingsRepository(sqliteDB.SQL())
	if err != nil {
		return fmt.Errorf("init settings repository: %w", err)
	}

	settingsService, err := user_settings.NewService(settingsRepo)
	if err != nil {
		return fmt.Errorf("init settings service: %w", err)
	}

	requestQueue := queue.NewRequestQueue()

	downloadService, err := download.NewService(
		settingsService,
		cobaltClient,
		requestQueue,
		cfg.TempDir,
		cfg.MaxFileBytes,
		cfg.DownloadTimeout,
	)
	if err != nil {
		return fmt.Errorf("init download service: %w", err)
	}
	_ = downloadService

	defaultSettings, err := settingsService.GetByUserID(context.Background(), 0)
	if err != nil {
		return fmt.Errorf("settings service smoke check failed: %w", err)
	}
	log.Info("services initialized", zap.Any("default_user_settings", defaultSettings))

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	<-ctx.Done()
	log.Info("shutdown signal received")
	return nil
}
