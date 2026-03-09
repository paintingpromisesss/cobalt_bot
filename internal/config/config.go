package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	TelegramBotToken        string
	CobaltBaseURL           string
	MaxFileBytes            int64
	DBPath                  string
	TempDir                 string
	RequestTimeout          time.Duration
	DownloadTimeout         time.Duration
	PickerSessionManagerTTL time.Duration
	LogLevel                string
}

func Load() (Config, error) {
	cfg := Config{
		TelegramBotToken: strings.TrimSpace(os.Getenv("TELEGRAM_BOT_TOKEN")),
		CobaltBaseURL:    getEnvDefault("COBALT_BASE_URL", "http://cobalt:9000/"),
		DBPath:           getEnvDefault("DB_PATH", "./data/bot.db"),
		TempDir:          getEnvDefault("TEMP_DIR", "./tmp"),
		LogLevel:         strings.ToLower(getEnvDefault("LOG_LEVEL", "info")),
	}

	if cfg.TelegramBotToken == "" {
		return Config{}, errors.New("TELEGRAM_BOT_TOKEN is required")
	}

	maxBytesRaw := strings.TrimSpace(os.Getenv("MAX_FILE_BYTES"))
	if maxBytesRaw == "" {
		return Config{}, errors.New("MAX_FILE_BYTES is required")
	} else {
		maxBytes, err := strconv.ParseInt(maxBytesRaw, 10, 64)
		if err != nil || maxBytes <= 0 {
			return Config{}, fmt.Errorf("MAX_FILE_BYTES must be a positive integer, got %q", maxBytesRaw)
		} else {
			cfg.MaxFileBytes = maxBytes
		}
	}

	requestTimeoutRaw := strings.TrimSpace(getEnvDefault("REQUEST_TIMEOUT", "30s"))
	requestTimeout, err := time.ParseDuration(requestTimeoutRaw)
	if err != nil || requestTimeout <= 0 {
		return Config{}, fmt.Errorf("REQUEST_TIMEOUT must be a positive duration, got %q", requestTimeoutRaw)
	} else {
		cfg.RequestTimeout = requestTimeout
	}

	downloadTimeoutRaw := strings.TrimSpace(getEnvDefault("DOWNLOAD_TIMEOUT", "10m"))
	downloadTimeout, err := time.ParseDuration(downloadTimeoutRaw)
	if err != nil || downloadTimeout <= 0 {
		return Config{}, fmt.Errorf("DOWNLOAD_TIMEOUT must be a positive duration, got %q", downloadTimeoutRaw)
	} else {
		cfg.DownloadTimeout = downloadTimeout
	}

	pickerSessionManangerTTLRaw := strings.TrimSpace(getEnvDefault("PICKER_SESSION_MANAGER_TTL", "10m"))
	pickerSessionManagerTTL, err := time.ParseDuration(pickerSessionManangerTTLRaw)
	if err != nil || pickerSessionManagerTTL <= 0 {
		return Config{}, fmt.Errorf("PICKER_SESSION_MANAGER_TTL must be a positive duration, got %q", pickerSessionManangerTTLRaw)
	} else {
		cfg.PickerSessionManagerTTL = pickerSessionManagerTTL
	}

	return cfg, nil
}

func getEnvDefault(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}
