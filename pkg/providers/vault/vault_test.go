package vault

import (
	"context"
	"github.com/camaeel/vault-autounseal-operator/pkg/config"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
	"time"
)

func TestGetVaultClusterNode(t *testing.T) {
	ctx := context.TODO()
	cfg := config.Config{
		TlsSkipVerify:        false,
		ServiceDomain:        "vault-internal.vault.svc.cluster.local",
		ServicePort:          8200,
		ServiceScheme:        "https",
		VaultTimeoutDuration: 10 * time.Second,
	}
	pod := corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: "vault-0",
		},
	}
	n, err := GetVaultClusterNode(ctx, &cfg, &pod)
	assert.NoError(t, err)
	assert.NotNil(t, n)
	assert.Equal(t, "https://vault-0.vault-internal.vault.svc.cluster.local:8200", n.Client.Address())
	assert.Equal(t, 10*time.Second, n.Client.ClientTimeout())
}
