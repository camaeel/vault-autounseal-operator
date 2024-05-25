package secrets

import (
	"context"
	"github.com/camaeel/vault-autounseal-operator/pkg/config"
	vault "github.com/hashicorp/vault/api"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/fake"
	"testing"
)

func TestCreateOrReplaceRootTokenSecretNew(t *testing.T) {
	cfg := config.Config{
		Namespace:            "vault",
		K8sClient:            fake.NewSimpleClientset(),
		VaultRootTokenSecret: "root-token-secret",
	}
	initData := vault.InitResponse{
		RootToken: "root-token-value",
	}

	err := CreateOrReplaceRootTokenSecret(&cfg, context.TODO(), &initData)
	assert.NoError(t, err)

	list, err := cfg.K8sClient.CoreV1().Secrets(v1.NamespaceAll).List(context.TODO(), v1.ListOptions{})
	assert.NoError(t, err)
	assert.NotNil(t, list)
	assert.Len(t, list.Items, 1)
	strData := list.Items[0].StringData
	assert.Equal(t, map[string]string{"rootToken": "root-token-value"}, strData)
}

func TestCreateOrReplaceRootTokenSecretExists(t *testing.T) {
	existingSecret := corev1.Secret{
		ObjectMeta: v1.ObjectMeta{
			Name:      "root-token-secret",
			Namespace: "vault",
		},
	}
	cfg := config.Config{
		Namespace:            "vault",
		K8sClient:            fake.NewSimpleClientset(&existingSecret),
		VaultRootTokenSecret: "root-token-secret",
	}
	initData := vault.InitResponse{
		RootToken: "root-token-value",
	}

	err := CreateOrReplaceRootTokenSecret(&cfg, context.TODO(), &initData)
	assert.NoError(t, err)

	list, err := cfg.K8sClient.CoreV1().Secrets(v1.NamespaceAll).List(context.TODO(), v1.ListOptions{})
	assert.NoError(t, err)
	assert.NotNil(t, list)
	assert.Len(t, list.Items, 1)
	strData := list.Items[0].StringData
	assert.Equal(t, map[string]string{"rootToken": "root-token-value"}, strData)
}

func TestCreateUnlockSecret(t *testing.T) {
	cfg := config.Config{
		Namespace:             "vault",
		K8sClient:             fake.NewSimpleClientset(),
		VaultUnlockKeysSecret: "unseal-secret",
	}
	initData := vault.InitResponse{
		Keys: []string{"key1", "key2", "key3"},
	}
	err := CreateUnlockSecret(&cfg, context.TODO(), &initData)
	assert.NoError(t, err)
	list, err := cfg.K8sClient.CoreV1().Secrets(v1.NamespaceAll).List(context.TODO(), v1.ListOptions{})
	assert.NoError(t, err)
	assert.NotNil(t, list)
	assert.Len(t, list.Items, 1)
	strData := list.Items[0].StringData
	assert.Equal(t, map[string]string{
		"unsealKey0": "key1",
		"unsealKey1": "key2",
		"unsealKey2": "key3",
	}, strData)
	assert.Equal(t, "unseal-secret", list.Items[0].Name)

	assert.Equal(t, "vault", list.Items[0].Namespace)
}

func TestGetUnlockSecretNotFound(t *testing.T) {
	cfg := config.Config{
		Namespace:             "vault",
		K8sClient:             fake.NewSimpleClientset(),
		VaultUnlockKeysSecret: "vault-unlock-secret",
	}

	fakeInformer := informers.NewSharedInformerFactoryWithOptions(cfg.K8sClient, 1, informers.WithNamespace(cfg.Namespace))
	secretInformerFactory := fakeInformer.Core().V1().Secrets()
	secretLister := secretInformerFactory.Lister()

	res, err := GetUnlockSecret(&cfg, secretLister)
	assert.True(t, errors.IsNotFound(err))
	assert.Nil(t, res)
}

func TestGetUnlockSecretFound(t *testing.T) {
	ctx := context.TODO()
	cfg := config.Config{
		Namespace: "vault",
		K8sClient: fake.NewSimpleClientset(
			&corev1.Secret{
				ObjectMeta: v1.ObjectMeta{
					Name:      "vault-unlock-secret",
					Namespace: "vault",
				},
				StringData: map[string]string{
					"unsealKey0": "key1",
				},
			},
		),
		VaultUnlockKeysSecret: "vault-unlock-secret",
	}

	fakeInformer := informers.NewSharedInformerFactoryWithOptions(cfg.K8sClient, 1, informers.WithNamespace(cfg.Namespace))
	secretInformerFactory := fakeInformer.Core().V1().Secrets()
	secretLister := secretInformerFactory.Lister()

	fakeInformer.Start(ctx.Done())
	fakeInformer.WaitForCacheSync(ctx.Done())

	res, err := GetUnlockSecret(&cfg, secretLister)
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, "key1", res.StringData["unsealKey0"])
	assert.Equal(t, "vault-unlock-secret", res.Name)
	assert.Len(t, res.StringData, 1)
}
