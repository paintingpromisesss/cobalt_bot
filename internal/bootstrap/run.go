package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"os/signal"
	"syscall"

	"github.com/paintingpromisesss/cobalt_bot/internal/platform/config"
	"go.uber.org/zap"
)

func Run(cfg config.Config) error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	container, err := Build(ctx, cfg)
	if err != nil {
		return err
	}
	defer closeContainer(container)

	logStartupConfig(container.Logger, cfg)

	if err := container.Handler.RegisterHandlers(); err != nil {
		container.Logger.Error("register handlers failed", zap.Error(err))
		return fmt.Errorf("register handlers: %w", err)
	}

	container.Bot.Start()
	<-ctx.Done()
	container.Logger.Info("shutdown signal received")
	return nil
}

func closeContainer(container *Container) {
	if container == nil {
		return
	}

	if container.Bot != nil {
		container.Bot.Stop()
	}

	if container.Storage != nil {
		if err := container.Storage.Close(); err != nil && container.Logger != nil {
			container.Logger.Warn("db close failed", zap.Error(err))
		}
	}

	if container.Logger != nil {
		if err := container.Logger.Sync(); err != nil &&
			!errors.Is(err, syscall.ENOTTY) &&
			!errors.Is(err, syscall.EINVAL) {
			container.Logger.Warn("logger sync failed", zap.Error(err))
		}
	}
}

func logStartupConfig(log *zap.Logger, cfg config.Config) {
	if log == nil {
		return
	}

	log.Info(
		"config loaded",
		zap.String("telegram_bot_api_url", cfg.Telegram.BotAPIURL),
		zap.Bool("telegram_local_file_mode", cfg.Telegram.LocalFileMode),
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

	if cfg.Telegram.LocalFileMode {
		log.Warn(
			"telegram local file mode is enabled; make sure the bot was logged out from the cloud Bot API before the first stable run against the local server",
		)
	}
}
