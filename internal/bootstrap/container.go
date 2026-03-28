package bootstrap

import (
	"context"
	"fmt"

	"github.com/paintingpromisesss/cobalt_bot/internal/adapters/cobalt"
	"github.com/paintingpromisesss/cobalt_bot/internal/adapters/fetch"
	"github.com/paintingpromisesss/cobalt_bot/internal/adapters/memory"
	"github.com/paintingpromisesss/cobalt_bot/internal/adapters/sqlite"
	"github.com/paintingpromisesss/cobalt_bot/internal/adapters/urlpolicy"
	"github.com/paintingpromisesss/cobalt_bot/internal/adapters/ytdlp"
	"github.com/paintingpromisesss/cobalt_bot/internal/platform/config"
	"github.com/paintingpromisesss/cobalt_bot/internal/platform/logger"
	"github.com/paintingpromisesss/cobalt_bot/internal/telegram"
	"github.com/paintingpromisesss/cobalt_bot/internal/telegram/handlers"
	"github.com/paintingpromisesss/cobalt_bot/internal/telegram/media"
	usecasedownload "github.com/paintingpromisesss/cobalt_bot/internal/usecase/download"
	usecasepicker "github.com/paintingpromisesss/cobalt_bot/internal/usecase/picker"
	usecasesettings "github.com/paintingpromisesss/cobalt_bot/internal/usecase/settings"
	usecasestart "github.com/paintingpromisesss/cobalt_bot/internal/usecase/start"
	"go.uber.org/zap"
)

type Container struct {
	Logger  *zap.Logger
	Storage *sqlite.DB
	Bot     *telegram.Bot
	Handler *handlers.Handler
}

func Build(ctx context.Context, cfg config.Config) (*Container, error) {
	log, err := logger.New(cfg.Logging.Level)
	if err != nil {
		return nil, fmt.Errorf("init logger: %w", err)
	}

	db, err := sqlite.New(cfg.Storage.DBPath)
	if err != nil {
		_ = log.Sync()
		return nil, fmt.Errorf("init db: %w", err)
	}

	tgBot, err := telegram.New(cfg.Telegram.BotToken, cfg.Telegram.BotAPIURL, cfg.Timeouts.TelegramSend, log)
	if err != nil {
		_ = db.Close()
		_ = log.Sync()
		return nil, fmt.Errorf("init telegram bot: %w", err)
	}

	queueManager := memory.NewUserJobGuard()
	cobaltClient := cobalt.NewClient(cfg.Cobalt.BaseURL, cfg.Timeouts.Request)
	cobaltGateway := cobalt.NewDownloadGateway(cobaltClient)
	fileDownloader := fetch.NewDownloader(cfg.Timeouts.Download, cfg.Storage.TempDir, cfg.Storage.MaxFileBytes)
	ytDLPClient := ytdlp.NewClient(
		cfg.Storage.TempDir,
		cfg.YTDLP.MaxMediaDurationSeconds,
		cfg.Storage.MaxFileBytes,
		cfg.YTDLP.CurrentlyLiveAvailable,
		cfg.YTDLP.PlaylistAvailable,
		cfg.YTDLP.UseJSRuntime,
	)
	ytDLPGateway := ytdlp.NewDownloadGateway(ytDLPClient)
	mediaSender := media.NewSender(log, cfg.Timeouts.FFprobe, cfg.Timeouts.FFmpeg, cfg.Telegram.LocalFileMode)
	mediaService := media.NewService(log, fileDownloader, ytDLPClient, mediaSender)

	instanceInfo, err := cobaltClient.GetInstanceInfo(ctx)
	if err != nil {
		_ = db.Close()
		_ = log.Sync()
		return nil, fmt.Errorf("get instance info: %w", err)
	}

	availableServices := instanceInfo.Cobalt.Services
	urlValidator := urlpolicy.NewURLValidator(availableServices)
	pickerSessionManager := memory.NewPickerStore(ctx, cfg.PickerSession.TTL, cfg.PickerSession.CleanupInterval)
	settingsService := usecasesettings.NewService(db, sqlite.ErrUserSettingsNotFound)
	downloadService := usecasedownload.NewService(
		settingsService,
		urlValidator,
		cobaltGateway,
		ytDLPGateway,
	)
	pickerService := usecasepicker.NewService(pickerSessionManager)
	startService := usecasestart.NewService(settingsService, availableServices)

	handler := handlers.NewHandler(
		ctx,
		cfg.Timeouts.Request,
		cfg.Timeouts.Download,
		cfg.YTDLP.MaxMediaDurationSeconds,
		tgBot,
		queueManager,
		log,
		mediaService,
		downloadService,
		pickerService,
		startService,
	)

	return &Container{
		Logger:  log,
		Storage: db,
		Bot:     tgBot,
		Handler: handler,
	}, nil
}
