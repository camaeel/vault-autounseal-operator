package podhandler

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	vaultProvider "github.com/camaeel/vault-autounseal-operator/pkg/providers/vault"
	"golang.org/x/exp/maps"
	"k8s.io/apimachinery/pkg/api/errors"

	"github.com/camaeel/vault-autounseal-operator/pkg/config"
	operatorSecrets "github.com/camaeel/vault-autounseal-operator/pkg/vault-autounseal-operator/secrets"
	corev1 "k8s.io/api/core/v1"
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
		err = initialize(logger, ctx, cfg, secretLister, vaultNode)
		if err != nil {
			logger.Error(fmt.Sprintf("Can't initialize vault node due to: %v", err))
		}
	} else if sealed {
		err = unseal(logger, ctx, cfg, secretLister, vaultNode)
		if err != nil {
			logger.Error(fmt.Sprintf("Can't unseal vault node due to: %v", err))
		}
	}

	// check if certificate served by vault doesn't match one in secret
	//// drain pod (so the API will keep minimum pods according to PDB)

}

func initialize(logger *slog.Logger, ctx context.Context, cfg *config.Config, secretLister listerv1.SecretLister, vaultNode vaultProvider.Node) error {
	mutex.Lock()
	defer mutex.Unlock()

	logger.Info("Pod not initialized. Attempting initialization")
	initSecret, err := operatorSecrets.GetUnlockSecret(cfg, secretLister)
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("can't get vault initialization secret: %v", err)
	}
	if err == nil {
		if initSecret.CreationTimestamp.Add(cfg.InformerResync * 3).Before(time.Now()) {
			//secret is older than 3 informer resyncs - this shouldn't happen
			return fmt.Errorf("this pod isn't initialized yet, but initialization secret %s already exists and is older than %s - either this secret is old (from previous initialization) or initialization procedure failed", cfg.VaultUnlockKeysSecret, (3 * cfg.InformerResync).String())
		} else {
			logger.Warn(fmt.Sprintf("fmt.Sprintf(\"This vault pod is not yet initialized but initialization data secret: %s already exists and was created less than %s - probably vault is not yet fully initialized", cfg.VaultUnlockKeysSecret, (3 * cfg.InformerResync).String()))
			return nil
		}
	}

	unsealKeys, rootToken, err := vaultNode.Initialize(cfg, ctx)
	if err != nil {
		return fmt.Errorf("can't initialize vault: %v", err)
	}
	logger.Info("Pod initialized")

	err = operatorSecrets.CreateUnlockSecret(cfg, ctx, unsealKeys)
	if err != nil {
		return fmt.Errorf("can't create vault initialization secret: %v", err)

	}
	logger.Info("Init data secret created", "secret", cfg.VaultUnlockKeysSecret)
	logger.Info("Attempting to create root token secret", "secret", cfg.VaultRootTokenSecret)
	err = operatorSecrets.CreateOrReplaceRootTokenSecret(cfg, ctx, rootToken)
	if err != nil {
		return fmt.Errorf("can't create root token secret: %v", err)
	}
	logger.Info("Root token secret created", "secret", cfg.VaultRootTokenSecret)
	return nil
}

func unseal(logger *slog.Logger, ctx context.Context, cfg *config.Config, secretLister listerv1.SecretLister, vaultNode vaultProvider.Node) error {
	mutex.Lock()
	defer mutex.Unlock()

	logger.Info("Pod is sealed")
	initSecret, err := operatorSecrets.GetUnlockSecret(cfg, secretLister)
	if err != nil {
		return fmt.Errorf("can't usneal because, can't get vault initialization secret: %v", err)

	}
	err = vaultNode.Unseal(logger, ctx, maps.Values(initSecret.Data))

	if err != nil {
		return fmt.Errorf("can't unseal vault node: %v", err)
	}
	logger.Info("Pod has been unsealed")
	return nil
}
