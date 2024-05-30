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
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	listerv1 "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
)

var mutex = &sync.RWMutex{}

func GetPodHandlerFunctions(cfg *config.Config, ctx context.Context, secretLister listerv1.SecretLister) cache.ResourceEventHandlerFuncs {
	ret := cache.ResourceEventHandlerFuncs{
		AddFunc: func(newObj interface{}) {
			podHandler(cfg, ctx, secretLister, newObj.(*corev1.Pod))
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			podHandler(cfg, ctx, secretLister, newObj.(*corev1.Pod))
		},
		DeleteFunc: nil,
	}
	return ret

}

func podHandler(cfg *config.Config, ctx2 context.Context, secretLister listerv1.SecretLister, pod *corev1.Pod) {
	logger := slog.With(slog.String("pod", pod.Name))
	ctx, cancel := context.WithTimeout(ctx2, cfg.HandlerTimeoutDuration)
	defer cancel()
	logger.Debug("Starting pod handler")

	vaultNode, err := vaultProvider.GetVaultClusterNode(ctx, cfg, pod)
	if err != nil {
		logger.Error(fmt.Sprintf("Can't get vault client for pod, due to: %v", err))
		return
	}

	sealed, initialized, err := vaultNode.GetSealStatus(ctx)
	if !initialized {
		locked := mutex.TryLock()
		if locked {
			defer mutex.Unlock()
		} else {
			logger.Warn("can't obtain Write lock. Probably initialization in progress")
			return
		}

		logger.Info("Pod not initialized. Attempting initialization")
		_, err := operatorSecrets.GetUnlockSecret(cfg, secretLister)
		if err != nil && !errors.IsNotFound(err) {
			logger.Error(fmt.Sprintf("can't get vault initialization secret: %v", err))
			return
		}
		if err == nil {
			logger.Error(fmt.Sprintf("Initialization data secret: %s already exists, but the cluster is not yet initialized. Probably an error. Delete secret %s in namespace %s, and try again.", cfg.VaultUnlockKeysSecret, cfg.VaultUnlockKeysSecret, cfg.Namespace))
			return
		}

		unsealKeys, rootToken, err := vaultNode.Initialize(cfg, ctx)
		if err != nil {
			logger.Error(fmt.Sprintf("can't initialize vault: %v", err))
			return
		}
		logger.Info("Pod initialized")

		err = operatorSecrets.CreateUnlockSecret(cfg, ctx, unsealKeys)
		if err != nil {
			logger.Error(fmt.Sprintf("can't create vault initialization secret: %v", err))
			return
		}
		logger.Info("Init data secret created", "secret", cfg.VaultUnlockKeysSecret)
		logger.Info("Attempting to create root token secret", "secret", cfg.VaultRootTokenSecret)
		err = operatorSecrets.CreateOrReplaceRootTokenSecret(cfg, ctx, rootToken)
		if err != nil {
			logger.Error(fmt.Sprintf("can't create root token secret: %v", err))
			return
		}
		logger.Info("Root token secret created", "secret", cfg.VaultRootTokenSecret)
	} else if sealed {
		mutex.Lock()
		defer mutex.Unlock()

		logger.Info("Pod is sealed")
		initSecret, err := operatorSecrets.GetUnlockSecret(cfg, secretLister)
		if err != nil {
			logger.Error(fmt.Sprintf("can't usneal because, can't get vault initialization secret: %v", err))
			return
		}
		err = vaultNode.Unseal(logger, ctx, maps.Values(initSecret.Data))

		if err != nil {
			logger.Error(fmt.Sprintf("can't unseal vault: %v", err), "pod", pod.Name)
			return
		}
		logger.Info("Pod has been unsealed")
		return
	}

	// check if certificate served by vault doesn't match one in secret
	//// drain pod (so the API will keep minimum pods according to PDB)

}
