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
	podhandler "github.com/camaeel/vault-autounseal-operator/pkg/vault-autounseal-operator/pod_handler"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
)

func Exec(ctx context.Context, cfg *config.Config) error {
	slog.Info(fmt.Sprintf("Staring now %s", "vault-autounseal-operator"))

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
			// pod ifnroremr
			podTweakListOptionsFunc := func(opts *v1.ListOptions) {
				opts.LabelSelector = cfg.PodSelector
			}
			//stsTweakListOptionsFunc := func(opts *v1.ListOptions) {
			//	opts.LabelSelector = cfg.StatefulsetSelector
			//
			//}
			podFactory := informers.NewSharedInformerFactoryWithOptions(cfg.K8sClient, cfg.InformerResync,
				informers.WithNamespace(cfg.Namespace),
				informers.WithTweakListOptions(podTweakListOptionsFunc),
			)
			//stsFactory := informers.NewSharedInformerFactoryWithOptions(cfg.K8sClient, cfg.InformerResync,
			//	informers.WithNamespace(cfg.Namespace),
			//	informers.WithTweakListOptions(stsTweakListOptionsFunc),
			//)
			secretFactory := informers.NewSharedInformerFactoryWithOptions(cfg.K8sClient, cfg.InformerResync,
				informers.WithNamespace(cfg.Namespace),
				// informers.WithTweakListOptions(podTweakListOptionsFunc),
			)

			podInformerFactory := podFactory.Core().V1().Pods()
			//stsInformerFactory := stsFactory.Apps().V1().StatefulSets()
			secretInformerFactory := secretFactory.Core().V1().Secrets()

			podInformer := podInformerFactory.Informer()
			//stsInformer := stsInformerFactory.Informer()
			secretInformer := secretInformerFactory.Informer()

			secretLister := secretInformerFactory.Lister()

			podFactory.Start(ctx.Done())
			podFactory.WaitForCacheSync(ctx.Done())
			//stsFactory.Start(ctx.Done())
			//stsFactory.WaitForCacheSync(ctx.Done())
			secretFactory.Start(ctx.Done())
			secretFactory.WaitForCacheSync(ctx.Done())

			_, err := podInformer.AddEventHandler(podhandler.GetPodHandlerFunctions(cfg, ctx, secretLister))
			if err != nil {
				slog.Error("Failed to add event handler: %v", err)
				cancel()
			}

			if !cache.WaitForCacheSync(ctx.Done(), podInformer.HasSynced) {
				slog.Error("Timed out waiting for caches to sync")
				cancel()
			}
			//if !cache.WaitForCacheSync(ctx.Done(), stsInformer.HasSynced) {
			//	slog.Error("Timed out waiting for caches to sync")
			//	cancel()
			//}
			if !cache.WaitForCacheSync(ctx.Done(), secretInformer.HasSynced) {
				slog.Error("Timed out waiting for caches to sync")
				cancel()
			}

			// _, err = stsInformer.AddEventHandler(stsHandler.GetStsHandlerFunctions())
			// if err != nil {
			// 	slog.Error("Failed to add event handler: %v", err)
			// 	cancel()
			// }
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
