package podhandler

import (
	"context"
	"log/slog"
	"sync"

	"github.com/camaeel/vault-autounseal-operator/pkg/config"
	vaultProvider "github.com/camaeel/vault-autounseal-operator/pkg/providers/vault"
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

			initData, err := initialize(cfg, ctx, obj.(corev1.Pod))
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
		err := unseal(
			obj.(corev1.Pod),
		)
		if err != nil {
			slog.Error("can't unseal vault: %v", err)
		}
		return
	}

	// check if certificate served by vault doesn't match one in secret
	//// drain pod (so the API will keep minimum pods according to PDB)
	// check if cluster is initialized
	//// if there is no vault secret
	////// initialize
	// check if cluster is sealed
	//// unseal

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

func initialize(cfg *config.Config, ctx context.Context, pod corev1.Pod) (*vault.InitResponse, error) {
	req := vault.InitRequest{
		SecretShares:    cfg.UnlockShares,
		SecretThreshold: cfg.UnlockThreshold,
	}
	vaultClient := vaultProvider.GetVaultClient(pod)
	resp, err := vaultClient.Sys().InitWithContext(ctx, &req)
	if err != nil {
		slog.Error("Can't initialize vault")
	}
	return resp, err
}

func unseal(pod corev1.Pod) error {
	return nil
}

func drain(pod corev1.Pod) error {
	return nil
}
