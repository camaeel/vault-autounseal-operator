package secrets

import (
	"context"
	"fmt"
	"github.com/camaeel/vault-autounseal-operator/pkg/config"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	listerv1 "k8s.io/client-go/listers/core/v1"
	"log/slog"
)

func CreateOrReplaceRootTokenSecret(cfg *config.Config, ctx context.Context, rootToken string) error {
	secret := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: cfg.VaultRootTokenSecret,
		},
		StringData: map[string]string{
			"token": rootToken,
		},
	}
	_, err := cfg.K8sClient.CoreV1().Secrets(cfg.Namespace).Get(ctx, cfg.VaultRootTokenSecret, metav1.GetOptions{})
	if err != nil {
		if !errors.IsNotFound(err) {
			slog.Error("can't get root token secret")
			return err
		} else {
			_, err = cfg.K8sClient.CoreV1().Secrets(cfg.Namespace).Create(ctx, &secret, metav1.CreateOptions{})
			return err
		}
	}

	_, err = cfg.K8sClient.CoreV1().Secrets(cfg.Namespace).Update(ctx, &secret, metav1.UpdateOptions{})
	return err
}

func GetUnlockSecret(cfg *config.Config, secretLister listerv1.SecretLister) (*corev1.Secret, error) {
	return secretLister.Secrets(cfg.Namespace).Get(cfg.VaultUnlockKeysSecret)
}

func CreateUnlockSecret(cfg *config.Config, ctx context.Context, unsealKeys []string) error {

	secret := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: cfg.VaultUnlockKeysSecret,
		},
		StringData: map[string]string{},
	}
	for i := range unsealKeys {
		secret.StringData[fmt.Sprintf("unsealKey%d", i)] = unsealKeys[i]
	}

	_, err := cfg.K8sClient.CoreV1().Secrets(cfg.Namespace).Create(ctx, &secret, metav1.CreateOptions{})
	return err
}
