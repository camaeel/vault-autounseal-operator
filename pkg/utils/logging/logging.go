package logging

import (
	"errors"
	"log/slog"
	"os"

	"github.com/camaeel/vault-autounseal-operator/pkg/config"
)

func SetupLogging(cfg *config.Config) error {
	var logger *slog.Logger

	var logLevel slog.Level
	switch cfg.LogLevel {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	case "":
		logLevel = slog.LevelInfo
	default:
		return errors.New("invalid log level")
	}

	if cfg.LogFormat == "text" {
		logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: logLevel,
		}))

	} else if cfg.LogFormat == "json" {
		logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: logLevel,
		}))
	} else {
		return errors.New("invalid log format")
	}

	slog.SetDefault(logger)
	return nil
}
