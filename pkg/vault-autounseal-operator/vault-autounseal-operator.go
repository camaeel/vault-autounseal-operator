package vaultAutounsealOperator

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/camaeel/vault-autounseal-operator/pkg/config"
	"github.com/camaeel/vault-autounseal-operator/pkg/providers/kubeclient"
	"github.com/camaeel/vault-autounseal-operator/pkg/utils/logger"
	podhandler "github.com/camaeel/vault-autounseal-operator/pkg/vault-autounseal-operator/pod_handler"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
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

	//TODO: implmenet filewatcher that will watch changes of ca.crt provided to the app

	kubeclient.LeaderElection(
		ctx,
		cfg,
		func(ctx3 context.Context) {
			factory := informers.NewSharedInformerFactoryWithOptions(cfg.K8sClient, cfg.InformerResync)
			podInformer := factory.Core().V1().Pods()
			informer := podInformer.Informer()
			// podLister := podInformer.Lister()
			factory.Start(ctx.Done())
			factory.WaitForCacheSync(ctx.Done())
			if !cache.WaitForCacheSync(ctx.Done(), informer.HasSynced) {
				slog.Error("Timed out waiting for caches to sync")
				cancel()
			}
			_, err := informer.AddEventHandler(podhandler.GetPodHandlerFunctions())
			if err != nil {
				slog.Error("Failed to add event handler: %v", err)
				cancel()
			}
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
