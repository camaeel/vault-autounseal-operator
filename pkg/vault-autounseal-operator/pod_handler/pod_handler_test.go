package podhandler

import (
	"context"
	"github.com/camaeel/vault-autounseal-operator/pkg/config"
	vaultProvider "github.com/camaeel/vault-autounseal-operator/pkg/providers/vault"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/fake"
	"log/slog"
	"testing"
	"time"
)

func TestGetPodHandlerFunctions(t *testing.T) {
	cfg := config.Config{
		Namespace: "vault",
	}
	ret := GetPodHandlerFunctions(&cfg, context.TODO(), nil)
	assert.NotNil(t, ret)
}

func TestInitialize(t *testing.T) {
	ctx := context.TODO()
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: "vault-0",
		},
	}

	fakeVault := vaultProvider.GetVault(t, false, 3)
	cfg := config.Config{
		Namespace: "vault",
		PodAddressesMap: map[string]string{
			"vault-0": fakeVault[0].Client.Address(),
		},
		ServiceScheme:          "https",
		TlsSkipVerify:          true,
		HandlerTimeoutDuration: 30 * time.Second,
		VaultTimeoutDuration:   10 * time.Second,
		UnlockShares:           3,
		UnlockThreshold:        2,
		K8sClient:              fake.NewSimpleClientset(),
		VaultUnlockKeysSecret:  "vault-autounseal-unlock-keys",
		VaultRootTokenSecret:   "vault-autounseal-root-token",
	}

	fakeClient := fake.NewSimpleClientset() //no secret
	fakeInformer := informers.NewSharedInformerFactoryWithOptions(fakeClient, 1, informers.WithNamespace(cfg.Namespace))
	secretInformerFactory := fakeInformer.Core().V1().Secrets()
	fakeSecretLister := secretInformerFactory.Lister()

	vaultNode, err := vaultProvider.GetVaultClusterNode(ctx, &cfg, pod)
	assert.NoError(t, err)

	err = initialize(slog.Default(), ctx, &cfg, fakeSecretLister, vaultNode)
	assert.NoError(t, err)

	sealed, initialized, err := vaultNode.GetSealStatus(ctx)

	assert.NoError(t, err)
	assert.True(t, initialized)
	assert.True(t, sealed)

	unlockSecret, err := cfg.K8sClient.CoreV1().Secrets(cfg.Namespace).Get(ctx, cfg.VaultUnlockKeysSecret, metav1.GetOptions{})
	assert.NoError(t, err)
	assert.Len(t, unlockSecret.StringData, cfg.UnlockShares) //probably "feature" of fake client - check StringData instead of Data

	rootSecret, err := cfg.K8sClient.CoreV1().Secrets(cfg.Namespace).Get(ctx, cfg.VaultRootTokenSecret, metav1.GetOptions{})
	assert.NoError(t, err)
	assert.Len(t, rootSecret.StringData, 1) //probably "feature" of fake client - check StringData instead of Data
}

//func TestPodHandlerUnseal(t *testing.T) {
//	ctx := context.TODO()
//	pod := &corev1.Pod{
//		ObjectMeta: metav1.ObjectMeta{
//			Name: "vault-0",
//		},
//	}
//
//	fakeVault := vaultProvider.GetVault(t, false, 3)
//	cfg := config.Config{
//		Namespace: "vault",
//		PodAddressesMap: map[string]string{
//			"vault-0": fakeVault[0].Client.Address(),
//		},
//		ServiceScheme:          "https",
//		TlsSkipVerify:          true,
//		VaultUnlockKeysSecret:  "vault-autounseal-unlock-keys",
//		VaultRootTokenSecret:   "vault-autounseal-root-token",
//		HandlerTimeoutDuration: 30 * time.Second,
//		VaultTimeoutDuration:   10 * time.Second,
//		UnlockShares:           3,
//		UnlockThreshold:        2,
//		K8sClient:              fake.NewSimpleClientset(),
//	}
//
//	node, err := vaultProvider.GetVaultClusterNode(ctx, &cfg, pod)
//	assert.NoError(t, err)
//
//	unsealKeys, _, err := node.Initialize(&cfg, ctx)
//	assert.NoError(t, err)
//
//	//map unseal keys
//	unsealKeysMap := map[string]string{}
//	for i := range unsealKeys {
//		unsealKeysMap[fmt.Sprintf("key%d", i)] = unsealKeys[i]
//	}
//
//	fakeClient := fake.NewSimpleClientset(
//		&corev1.Secret{
//			ObjectMeta: metav1.ObjectMeta{
//				Name:      cfg.VaultUnlockKeysSecret,
//				Namespace: cfg.Namespace,
//			},
//			StringData: unsealKeysMap,
//		},
//	)
//	fakeInformer := informers.NewSharedInformerFactoryWithOptions(fakeClient, 1, informers.WithNamespace(cfg.Namespace))
//	secretInformerFactory := fakeInformer.Core().V1().Secrets()
//	fakeSecretLister := secretInformerFactory.Lister()
//
//	podHandler(&cfg, ctx, fakeSecretLister, pod)
//
//	sealed, initialized, err := node.GetSealStatus(ctx)
//
//	assert.NoError(t, err)
//	assert.True(t, initialized)
//	assert.False(t, sealed)
//}
