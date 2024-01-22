package logger

import (
	"log/slog"
	"os"

	"github.com/camaeel/vault-unseal-operator/pkg/config"
)

var logger *slog.Logger

func Logger() *slog.Logger {
	return logger
}

func SetupLogging(cfg *config.Config) {
	if cfg.LogFormat == "text" {
		logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))
	} else {
		logger = slog.Default()
	}
}
