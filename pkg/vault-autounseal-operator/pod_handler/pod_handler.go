package podhandler

import (
	"context"
	"fmt"
	vaultProvider "github.com/camaeel/vault-autounseal-operator/pkg/providers/vault"
	"golang.org/x/exp/maps"
	"log/slog"
	"sync"

	"github.com/camaeel/vault-autounseal-operator/pkg/config"
	operatorSecrets "github.com/camaeel/vault-autounseal-operator/pkg/vault-autounseal-operator/secrets"
	vault "github.com/hashicorp/vault/api"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	listerv1 "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
)

var mutex = &sync.RWMutex{}

func GetPodHandlerFunctions(cfg *config.Config, ctx context.Context, secretLister listerv1.SecretLister) cache.ResourceEventHandlerFuncs {
	ret := cache.ResourceEventHandlerFuncs{
		AddFunc: func(newObj interface{}) {
			podHandler(cfg, ctx, secretLister, newObj)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			podHandler(cfg, ctx, secretLister, newObj)
		},
		DeleteFunc: nil,
	}
	return ret

}

func podHandler(cfg *config.Config, ctx context.Context, secretLister listerv1.SecretLister, obj interface{}) {
	mutex.RLock()
	defer mutex.RUnlock()

	slog.Debug("Starting pod handler", "pod", obj.(*corev1.Pod).Name)

	vaultClient, err := vaultProvider.GetVaultClient(cfg, obj.(corev1.Pod))
	if err != nil {
		slog.Error(fmt.Sprintf("Can't get vault client for pod, due to: %v", err))
		return
	}

	initSecret, err := operatorSecrets.GetUnlockSecret(cfg, secretLister)
	if err != nil {
		if !errors.IsNotFound(err) {
			slog.Error("can't get vault initialization secret: %v", err)
			return
		}
		initSecret = nil
	}

	if !isInitialized(obj.(corev1.Pod)) {
		if initSecret != nil {
			mutex.Lock()
			defer mutex.Unlock()

			initData, err := initialize(cfg, ctx, vaultClient)
			if err != nil {
				slog.Error("can't initialize vault: %v", err)
				return
			}

			err = operatorSecrets.CreateUnlockSecret(cfg, ctx, initData)
			if err != nil {
				slog.Error("can't create vault initialization secret: %v", err)
				return
			}
			//how to handle if this fails. Then cluster is initialized but operator doesn't have initialization data. Cluster needs to be cleaned and initialized from scratch - might be detected and at least suggested in the logs.
			err = operatorSecrets.CreateOrReplaceRootTokenSecret(cfg, ctx, initData)
			if err != nil {
				slog.Error("can't create root token secret: %v", err)
			}
		}
		return //this should trigger another call to this method wit initialized=true
	}

	if isSealed(obj.(corev1.Pod)) {
		if initSecret == nil {
			slog.Error("init secret is not initialized, so can't unseal")
			return
		}
		err := unseal(ctx, vaultClient, maps.Values(initSecret.StringData)) //TODO: here
		if err != nil {
			slog.Error("can't unseal vault: %v", err)
		}
		return
	}

	// check if certificate served by vault doesn't match one in secret
	//// drain pod (so the API will keep minimum pods according to PDB)

}

func isInitialized(pod corev1.Pod) bool {
	return pod.Annotations["vault-initialized"] == "true"
}

func isSealed(pod corev1.Pod) bool {
	return pod.Annotations["vault-sealed"] == "true"
}

func isLeader(pod corev1.Pod) bool {
	return pod.Annotations["vault-active"] == "true"
}

func initialize(cfg *config.Config, ctx context.Context, vaultClient *vault.Client) (*vault.InitResponse, error) {
	req := vault.InitRequest{
		SecretShares:    cfg.UnlockShares,
		SecretThreshold: cfg.UnlockThreshold,
	}

	resp, err := vaultClient.Sys().InitWithContext(ctx, &req)
	if err != nil {
		slog.Error("Can't initialize vault")
	}
	return resp, err
}

func unseal(ctx context.Context, vaultClient *vault.Client, unsealData []string) error {
	for i := range unsealData {
		resp, err := vaultClient.Sys().UnsealWithContext(ctx, unsealData[i])
		if err != nil {
			return err
		}
		slog.Info(fmt.Sprintf("unseal resp: %v", resp))
	}
	return nil
}

func drain(pod corev1.Pod) error {
	return nil
}
