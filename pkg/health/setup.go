package health

import (
	"context"
	"fmt"
	"github.com/alexliesenfeld/health"
	"github.com/camaeel/vault-autounseal-operator/pkg/config"
	"log/slog"
	"net/http"
)

func Setup(ctx context.Context, cfg *config.Config) {
	checker := health.NewChecker(
		health.WithCheck(
			health.Check{
				Name: "liveness",
				// The check function checks the health of a component. If an error is
				// returned, the component is considered unavailable (or "down").
				// The context contains a deadline according to the configured timeouts.
				Check: func(ctx context.Context) error {
					return nil
				},
			}),
	)

	http.Handle("/healthz", health.NewHandler(checker))
	err := http.ListenAndServe(fmt.Sprintf(":%d", cfg.Port), nil)
	if err != nil {
		slog.Error(err.Error())
		panic(err)
	}
}
