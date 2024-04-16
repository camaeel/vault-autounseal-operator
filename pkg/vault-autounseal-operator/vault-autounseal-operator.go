package vaultAutounsealOperator

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/camaeel/vault-autounseal-operator/pkg/config"
	"github.com/camaeel/vault-autounseal-operator/pkg/kubeclient"
	"github.com/camaeel/vault-autounseal-operator/pkg/utils/logger"
)

func Exec(ctx context.Context, cfg *config.Config) error {
	logger.Logger().Info(fmt.Sprintf("Staring now %s asd", "asd"))

	clientset, currentNamespace, err := kubeclient.GetClient()
	if err != nil {
		return err
	}

	cfg.K8sClient = clientset
	if cfg.Namespace == "" {
		cfg.Namespace = currentNamespace
	}
	if cfg.LeaseNamespace == "" {
		cfg.LeaseNamespace = currentNamespace
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	cancelOnSigterm(cancel)

	kubeclient.LeaderElection(
		ctx,
		cfg,
		func(ctx3 context.Context) {
			fmt.Printf("DO the logic")
			//TODO informer
		},
		cancel)

	return nil
}

func cancelOnSigterm(cancel func()) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-ch
		slog.Info("Received termination, signaling shutdown")
		cancel()
	}()
}
