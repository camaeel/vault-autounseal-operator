package vaultClient

import (
	vault "github.com/hashicorp/vault/api"
	corev1 "k8s.io/api/core/v1"
)

func GetVaultClient(pod corev1.Pod) *vault.Client {

	return nil
}
