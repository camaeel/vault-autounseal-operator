package logging

import (
	"errors"
	"log/slog"
	"os"

	"github.com/camaeel/vault-autounseal-operator/pkg/config"
)

func SetupLogging(cfg *config.Config) error {
	var logger *slog.Logger
	if cfg.LogFormat == "text" {
		logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))

	} else if cfg.LogFormat == "json" {
		logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))
	} else {
		return errors.New("invalid log format")
	}
	slog.SetDefault(logger)
	return nil
}
