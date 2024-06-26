package podhandler

import (
	"context"
	"fmt"
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

	fakeVault := vaultProvider.GetVault(t, false, 3)
	cfg := config.Config{
		Namespace:              "vault",
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

	fakeInformer.Start(ctx.Done())
	fakeInformer.WaitForCacheSync(ctx.Done())

	vaultNode := vaultProvider.Node{Client: fakeVault[0].Client}

	pod := corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "vault-0",
			CreationTimestamp: metav1.NewTime(time.Now().Add(-1 * time.Second)),
		},
	}

	err := initialize(slog.Default(), ctx, &cfg, fakeSecretLister, vaultNode, &pod)
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

func TestInitializeFailOldInitSecretExistsAndOldPod(t *testing.T) {
	ctx := context.TODO()

	fakeVault := vaultProvider.GetVault(t, false, 3)
	cfg := config.Config{
		Namespace:              "vault",
		HandlerTimeoutDuration: 30 * time.Second,
		VaultTimeoutDuration:   10 * time.Second,
		UnlockShares:           3,
		UnlockThreshold:        2,
		K8sClient:              fake.NewSimpleClientset(),
		VaultUnlockKeysSecret:  "vault-autounseal-unlock-keys",
		VaultRootTokenSecret:   "vault-autounseal-root-token",
		InformerResync:         20 * time.Second,
	}

	fakeClient := fake.NewSimpleClientset(
		&corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:              cfg.VaultUnlockKeysSecret,
				Namespace:         cfg.Namespace,
				CreationTimestamp: metav1.NewTime(time.Now().Add(-1 * time.Hour)),
			},
			Data: map[string][]byte{},
		},
	)
	fakeInformer := informers.NewSharedInformerFactoryWithOptions(fakeClient, 1, informers.WithNamespace(cfg.Namespace))
	secretInformerFactory := fakeInformer.Core().V1().Secrets()
	fakeSecretLister := secretInformerFactory.Lister()

	fakeInformer.Start(ctx.Done())
	fakeInformer.WaitForCacheSync(ctx.Done())

	vaultNode := vaultProvider.Node{Client: fakeVault[0].Client}

	pod := corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "vault-0",
			CreationTimestamp: metav1.NewTime(time.Now().Add(-1 * time.Hour)),
		},
	}

	err := initialize(slog.Default(), ctx, &cfg, fakeSecretLister, vaultNode, &pod)
	assert.Error(t, err)
	assert.EqualError(t, err, fmt.Sprintf("this pod isn't initialized yet and is older than %s, but initialization secret vault-autounseal-unlock-keys already exists and is older than %s - either this secret is from previously initialized vault or initialization procedure failed", (RECONCILATION_LOOPS_TO_WAIT*cfg.InformerResync).String(), (RECONCILATION_LOOPS_TO_WAIT*cfg.InformerResync).String()))

	sealed, initialized, err := vaultNode.GetSealStatus(ctx)

	assert.NoError(t, err)
	assert.False(t, initialized)
	assert.True(t, sealed)
}

func TestInitializeFailRecentInitSecretExists(t *testing.T) {
	ctx := context.TODO()

	fakeVault := vaultProvider.GetVault(t, false, 3)
	cfg := config.Config{
		Namespace:              "vault",
		HandlerTimeoutDuration: 30 * time.Second,
		VaultTimeoutDuration:   10 * time.Second,
		UnlockShares:           3,
		UnlockThreshold:        2,
		K8sClient:              fake.NewSimpleClientset(),
		VaultUnlockKeysSecret:  "vault-autounseal-unlock-keys",
		VaultRootTokenSecret:   "vault-autounseal-root-token",
		InformerResync:         20 * time.Second,
	}

	fakeClient := fake.NewSimpleClientset(
		&corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:              cfg.VaultUnlockKeysSecret,
				Namespace:         cfg.Namespace,
				CreationTimestamp: metav1.NewTime(time.Now().Add(-1 * time.Second)),
			},
			Data: map[string][]byte{},
		},
	)
	fakeInformer := informers.NewSharedInformerFactoryWithOptions(fakeClient, 1, informers.WithNamespace(cfg.Namespace))
	secretInformerFactory := fakeInformer.Core().V1().Secrets()
	fakeSecretLister := secretInformerFactory.Lister()

	fakeInformer.Start(ctx.Done())
	fakeInformer.WaitForCacheSync(ctx.Done())

	vaultNode := vaultProvider.Node{Client: fakeVault[0].Client}

	pod := corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "vault-0",
			CreationTimestamp: metav1.NewTime(time.Now().Add(-1 * time.Hour)),
		},
	}

	err := initialize(slog.Default(), ctx, &cfg, fakeSecretLister, vaultNode, &pod)
	assert.NoError(t, err)

	sealed, initialized, err := vaultNode.GetSealStatus(ctx)

	assert.NoError(t, err)
	assert.False(t, initialized)
	assert.True(t, sealed)
}

func TestUnseal(t *testing.T) {
	ctx := context.TODO()

	fakeVault := vaultProvider.GetVault(t, false, 3)
	cfg := config.Config{
		Namespace:              "vault",
		VaultUnlockKeysSecret:  "vault-autounseal-unlock-keys",
		VaultRootTokenSecret:   "vault-autounseal-root-token",
		HandlerTimeoutDuration: 30 * time.Second,
		VaultTimeoutDuration:   10 * time.Second,
		UnlockShares:           3,
		UnlockThreshold:        2,
		K8sClient:              fake.NewSimpleClientset(),
	}

	vaultNode := vaultProvider.Node{
		Client: fakeVault[0].Client,
	}

	unsealKeys, _, err := vaultNode.Initialize(&cfg, ctx)
	assert.NoError(t, err)

	//map unseal keys
	unsealKeysMap := map[string][]byte{}
	for i := range unsealKeys {
		unsealKeysMap[fmt.Sprintf("key%d", i)] = []byte(unsealKeys[i])
	}

	fakeClient := fake.NewSimpleClientset(
		&corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      cfg.VaultUnlockKeysSecret,
				Namespace: cfg.Namespace,
			},
			Data: unsealKeysMap,
		},
	)
	fakeInformer := informers.NewSharedInformerFactoryWithOptions(fakeClient, 1, informers.WithNamespace(cfg.Namespace))
	secretInformerFactory := fakeInformer.Core().V1().Secrets()
	fakeSecretLister := secretInformerFactory.Lister()

	fakeInformer.Start(ctx.Done())
	fakeInformer.WaitForCacheSync(ctx.Done())

	//podHandler(&cfg, ctx, fakeSecretLister, pod)
	err = unseal(slog.Default(), ctx, &cfg, fakeSecretLister, vaultNode)
	assert.NoError(t, err)

	sealed, initialized, err := vaultNode.GetSealStatus(ctx)

	assert.NoError(t, err)
	assert.True(t, initialized)
	assert.False(t, sealed)
}

func TestPodHandlerUnsealFailedNoUnsealSecret(t *testing.T) {
	ctx := context.TODO()

	fakeVault := vaultProvider.GetVault(t, false, 3)
	cfg := config.Config{
		Namespace:              "vault",
		VaultUnlockKeysSecret:  "vault-autounseal-unlock-keys",
		VaultRootTokenSecret:   "vault-autounseal-root-token",
		HandlerTimeoutDuration: 30 * time.Second,
		VaultTimeoutDuration:   10 * time.Second,
		UnlockShares:           3,
		UnlockThreshold:        2,
		K8sClient:              fake.NewSimpleClientset(),
	}

	vaultNode := vaultProvider.Node{
		Client: fakeVault[0].Client,
	}

	unsealKeys, _, err := vaultNode.Initialize(&cfg, ctx)
	assert.NoError(t, err)

	//map unseal keys
	unsealKeysMap := map[string][]byte{}
	for i := range unsealKeys {
		unsealKeysMap[fmt.Sprintf("key%d", i)] = []byte(unsealKeys[i])
	}

	fakeClient := fake.NewSimpleClientset()

	fakeInformer := informers.NewSharedInformerFactoryWithOptions(fakeClient, 1, informers.WithNamespace(cfg.Namespace))
	secretInformerFactory := fakeInformer.Core().V1().Secrets()
	fakeSecretLister := secretInformerFactory.Lister()

	fakeInformer.Start(ctx.Done())
	fakeInformer.WaitForCacheSync(ctx.Done())

	//podHandler(&cfg, ctx, fakeSecretLister, pod)
	err = unseal(slog.Default(), ctx, &cfg, fakeSecretLister, vaultNode)
	assert.Error(t, err)
	assert.EqualError(t, err, "can't usneal because, can't get vault initialization secret: secret \"vault-autounseal-unlock-keys\" not found")

	sealed, initialized, err := vaultNode.GetSealStatus(ctx)

	assert.NoError(t, err)
	assert.True(t, initialized)
	assert.True(t, sealed)
}

func TestPodHandlerInitialize(t *testing.T) {
	ctx := context.TODO()

	pod := corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "vault-0",
			Namespace: "vault",
		},
	}

	fakeVault := vaultProvider.GetVault(t, false, 3)
	cfg := config.Config{
		Namespace:              "vault",
		HandlerTimeoutDuration: 30 * time.Second,
		VaultTimeoutDuration:   10 * time.Second,
		UnlockShares:           3,
		UnlockThreshold:        2,
		K8sClient:              fake.NewSimpleClientset(),
		VaultUnlockKeysSecret:  "vault-autounseal-unlock-keys",
		VaultRootTokenSecret:   "vault-autounseal-root-token",
		TlsSkipVerify:          true,
		ServiceScheme:          "https",
		ServiceDomain:          "vault-internal.vault.svc.cluster.local",
		PodAddressesMap: map[string]string{
			"vault-0": fakeVault[0].Client.Address(),
		},
	}

	fakeClient := fake.NewSimpleClientset() //no secret
	fakeInformer := informers.NewSharedInformerFactoryWithOptions(fakeClient, 1, informers.WithNamespace(cfg.Namespace))
	secretInformerFactory := fakeInformer.Core().V1().Secrets()
	fakeSecretLister := secretInformerFactory.Lister()

	fakeInformer.Start(ctx.Done())
	fakeInformer.WaitForCacheSync(ctx.Done())

	vaultNode := vaultProvider.Node{Client: fakeVault[0].Client}

	podHandler(&cfg, ctx, fakeSecretLister, &pod)

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

func TestPodHandlerUnseal(t *testing.T) {
	ctx := context.TODO()

	pod := corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "vault-0",
			Namespace: "vault",
		},
	}

	fakeVault := vaultProvider.GetVault(t, false, 3)
	cfg := config.Config{
		Namespace:              "vault",
		HandlerTimeoutDuration: 30 * time.Second,
		VaultTimeoutDuration:   10 * time.Second,
		UnlockShares:           3,
		UnlockThreshold:        2,
		K8sClient:              fake.NewSimpleClientset(),
		VaultUnlockKeysSecret:  "vault-autounseal-unlock-keys",
		VaultRootTokenSecret:   "vault-autounseal-root-token",
		TlsSkipVerify:          true,
		ServiceScheme:          "https",
		ServiceDomain:          "vault-internal.vault.svc.cluster.local",
		PodAddressesMap: map[string]string{
			"vault-0": fakeVault[0].Client.Address(),
		},
	}

	// initialize
	vaultNode := vaultProvider.Node{
		Client: fakeVault[0].Client,
	}
	initData, _, err := vaultNode.Initialize(&cfg, ctx)
	initSecretData := map[string][]byte{}
	for i := range initData {
		initSecretData[fmt.Sprintf("key%d", i)] = []byte(initData[i])
	}
	///

	fakeClient := fake.NewSimpleClientset(
		&corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      cfg.VaultUnlockKeysSecret,
				Namespace: "vault",
			},
			Data: initSecretData,
		},
	)
	fakeInformer := informers.NewSharedInformerFactoryWithOptions(fakeClient, 1, informers.WithNamespace(cfg.Namespace))
	secretInformerFactory := fakeInformer.Core().V1().Secrets()
	fakeSecretLister := secretInformerFactory.Lister()

	fakeInformer.Start(ctx.Done())
	fakeInformer.WaitForCacheSync(ctx.Done())

	podHandler(&cfg, ctx, fakeSecretLister, &pod)

	sealed, initialized, err := vaultNode.GetSealStatus(ctx)

	assert.NoError(t, err)
	assert.True(t, initialized)
	assert.False(t, sealed)
}
