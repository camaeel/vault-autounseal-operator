package vault

import (
	"github.com/camaeel/vault-autounseal-operator/pkg/config"
	"github.com/madflojo/testcerts"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func TestGetVaultClient(t *testing.T) {
	dir := t.TempDir()
	cert, _, err := testcerts.GenerateCertsToTempFile(dir)
	assert.NoError(t, err)

	cfg := config.Config{
		ServiceDomain: "vault-internal.vault.svc.cluster.local",
		CaCert:        cert,
		ServicePort:   8200,
	}
	pod := corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: "vault-0",
		},
		Spec: corev1.PodSpec{},
	}

	res, err := GetVaultClient(&cfg, &pod)
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, "https://vault-0.vault-internal.vault.svc.cluster.local:8200", res.Address())

}
