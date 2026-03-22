package logger

import (
	"fmt"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func New(level string) (*zap.Logger, error) {
	var atom zap.AtomicLevel
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "debug":
		atom = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	case "info", "":
		atom = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	case "warn", "warning":
		atom = zap.NewAtomicLevelAt(zapcore.WarnLevel)
	case "error":
		atom = zap.NewAtomicLevelAt(zapcore.ErrorLevel)
	default:
		return nil, fmt.Errorf("unsupported LOG_LEVEL %q, allowed: debug, info, warn, error", level)
	}

	cfg := zap.NewProductionConfig()
	cfg.Level = atom
	return cfg.Build()
}
