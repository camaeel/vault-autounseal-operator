package podhandler

import (
	"log/slog"
	"sync"

	"github.com/camaeel/vault-autounseal-operator/pkg/config"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	listerv1 "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
)

var mutex *sync.RWMutex = &sync.RWMutex{}

type InitData struct {
	RootToken  string
	UnsealKeys []string
}

func GetPodHandlerFunctions(cfg *config.Config, secretLister listerv1.SecretLister) cache.ResourceEventHandlerFuncs {
	ret := cache.ResourceEventHandlerFuncs{
		AddFunc: func(newObj interface{}) {
			podHandler(cfg, secretLister, newObj)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			podHandler(cfg, secretLister, newObj)
		},
		DeleteFunc: nil,
	}
	return ret

}

func podHandler(cfg *config.Config, secretLister listerv1.SecretLister, obj interface{}) {
	mutex.RLock()
	defer mutex.RUnlock()

	slog.Debug("Starting pod handler", "pod", obj.(*corev1.Pod).Name)

	initSecret, err := getInitializedSecret(cfg, secretLister)
	if err != nil {
		slog.Warn("can't get vault initialization secret: %v", err)
		return
	}

	if isInitialized(obj.(corev1.Pod)) && initSecret != nil {
		mutex.Lock()
		defer mutex.Unlock()

		initData, err := initialize(
			obj.(corev1.Pod),
		)
		if err != nil {
			slog.Error("can't initialize vault: %v", err)
		}

		err = createInitSecrets(cfg, initData)
		if err != nil {
			slog.Error("can't create vault initialization secret: %v", err)
		}
		//create initialization secret
		//create or replace root token secret

		return //this should trigger another call to this method wit initialzied=true
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

func initialize(pod corev1.Pod) ([]InitData, error) {
	return nil, nil
}

func unseal(pod corev1.Pod) error {
	return nil
}

func drain(pod corev1.Pod) error {
	return nil
}

func getInitializedSecret(cfg *config.Config, secretLister listerv1.SecretLister) (*v1.Secret, error) {
	ret, err := secretLister.Secrets(cfg.Namespace).Get(cfg.VaultUnlockKeysSecret)
	if err != nil {
		return nil, err
	}
	return ret, err
}

func createInitSecrets(cfg *config.Config, initData []InitData) error {
	return nil
}
