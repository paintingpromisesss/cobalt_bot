package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type TelegramConfig struct {
	BotToken  string
	BotAPIURL string
}

type CobaltConfig struct {
	BaseURL string
}

type StorageConfig struct {
	DBPath       string
	TempDir      string
	MaxFileBytes int64
}

type TimeoutsConfig struct {
	Request      time.Duration
	Download     time.Duration
	TelegramSend time.Duration
	FFprobe      time.Duration
	FFmpeg       time.Duration
}

type YTDLPConfig struct {
	MaxMediaDurationSeconds int
	CurrentlyLiveAvailable  bool
	PlaylistAvailable       bool
	UseJSRuntime            bool
}

type PickerSessionConfig struct {
	TTL             time.Duration
	CleanupInterval time.Duration
}

type LoggingConfig struct {
	Level string
}

type Config struct {
	Telegram      TelegramConfig
	Cobalt        CobaltConfig
	Storage       StorageConfig
	Timeouts      TimeoutsConfig
	YTDLP         YTDLPConfig
	PickerSession PickerSessionConfig
	Logging       LoggingConfig
}

func Load() (Config, error) {
	cfg := Config{
		Telegram: TelegramConfig{
			BotToken:  strings.TrimSpace(os.Getenv("TG_BOT_TOKEN")),
			BotAPIURL: strings.TrimSpace(getEnvDefault("TG_BOT_API_URL", "http://telegram-bot-api:8081")),
		},
		Cobalt: CobaltConfig{
			BaseURL: getEnvDefault("TG_BOT_COBALT_BASE_URL", "http://cobalt:9000/"),
		},
		Storage: StorageConfig{
			DBPath:  getEnvDefault("TG_BOT_DB_PATH", "./data/bot.db"),
			TempDir: getEnvDefault("TG_BOT_TEMP_DIR", "./tmp"),
		},
		Logging: LoggingConfig{
			Level: strings.ToLower(getEnvDefault("TG_BOT_LOG_LEVEL", "info")),
		},
	}

	if cfg.Telegram.BotToken == "" {
		return Config{}, errors.New("TG_BOT_TOKEN is required")
	}

	maxFileBytes, err := parsePositiveInt64Env("TG_BOT_MAX_FILE_BYTES")
	if err != nil {
		return Config{}, err
	}
	cfg.Storage.MaxFileBytes = maxFileBytes

	requestTimeout, err := parsePositiveDurationEnv("TG_BOT_REQUEST_TIMEOUT", "30s")
	if err != nil {
		return Config{}, err
	}
	cfg.Timeouts.Request = requestTimeout

	downloadTimeout, err := parsePositiveDurationEnv("TG_BOT_DOWNLOAD_TIMEOUT", "10m")
	if err != nil {
		return Config{}, err
	}
	cfg.Timeouts.Download = downloadTimeout

	telegramSendTimeout, err := parsePositiveDurationEnv("TG_BOT_TELEGRAM_SEND_TIMEOUT", "10m")
	if err != nil {
		return Config{}, err
	}
	cfg.Timeouts.TelegramSend = telegramSendTimeout

	ffprobeTimeout, err := parsePositiveDurationEnv("TG_BOT_FFPROBE_TIMEOUT", "5s")
	if err != nil {
		return Config{}, err
	}
	cfg.Timeouts.FFprobe = ffprobeTimeout

	ffmpegTimeout, err := parsePositiveDurationEnv("TG_BOT_FFMPEG_TIMEOUT", "30s")
	if err != nil {
		return Config{}, err
	}
	cfg.Timeouts.FFmpeg = ffmpegTimeout

	maxMediaDurationSeconds, err := parsePositiveIntEnv("TG_BOT_YTDLP_MAX_MEDIA_DURATION_SECS", "7200")
	if err != nil {
		return Config{}, err
	}
	cfg.YTDLP.MaxMediaDurationSeconds = maxMediaDurationSeconds

	currentlyLiveAvailable, err := parseBoolEnv("TG_BOT_YTDLP_CURRENTLY_LIVE_AVAILABLE", "0")
	if err != nil {
		return Config{}, err
	}
	cfg.YTDLP.CurrentlyLiveAvailable = currentlyLiveAvailable

	playlistAvailable, err := parseBoolEnv("TG_BOT_YTDLP_PLAYLIST_AVAILABLE", "0")
	if err != nil {
		return Config{}, err
	}
	cfg.YTDLP.PlaylistAvailable = playlistAvailable

	useJSRuntime, err := parseBoolEnv("TG_BOT_YTDLP_USE_JS_RUNTIME", "0")
	if err != nil {
		return Config{}, err
	}
	cfg.YTDLP.UseJSRuntime = useJSRuntime

	pickerSessionTTL, err := parsePositiveDurationEnv("TG_BOT_PICKER_SESSION_MANAGER_TTL", "10m")
	if err != nil {
		return Config{}, err
	}
	cfg.PickerSession.TTL = pickerSessionTTL

	pickerSessionCleanupInterval, err := parsePositiveDurationEnv("TG_BOT_PICKER_SESSION_MANAGER_CLEANUP_INTERVAL", "3m")
	if err != nil {
		return Config{}, err
	}
	cfg.PickerSession.CleanupInterval = pickerSessionCleanupInterval

	return cfg, nil
}

func parsePositiveInt64Env(key string) (int64, error) {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return 0, fmt.Errorf("%s is required", key)
	}

	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || value <= 0 {
		return 0, fmt.Errorf("%s must be a positive integer, got %q", key, raw)
	}

	return value, nil
}

func parsePositiveIntEnv(key, fallback string) (int, error) {
	raw := strings.TrimSpace(getEnvDefault(key, fallback))

	value, err := strconv.Atoi(raw)
	if err != nil || value <= 0 {
		return 0, fmt.Errorf("%s must be a positive integer, got %q", key, raw)
	}

	return value, nil
}

func parsePositiveDurationEnv(key, fallback string) (time.Duration, error) {
	raw := strings.TrimSpace(getEnvDefault(key, fallback))

	value, err := time.ParseDuration(raw)
	if err != nil || value <= 0 {
		return 0, fmt.Errorf("%s must be a positive duration, got %q", key, raw)
	}

	return value, nil
}

func parseBoolEnv(key, fallback string) (bool, error) {
	raw := strings.TrimSpace(getEnvDefault(key, fallback))

	value, err := strconv.ParseBool(raw)
	if err == nil {
		return value, nil
	}

	switch raw {
	case "0":
		return false, nil
	case "1":
		return true, nil
	default:
		return false, fmt.Errorf("%s must be a boolean, got %q", key, raw)
	}
}

func getEnvDefault(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}
