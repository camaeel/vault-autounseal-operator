package vault

import (
	"fmt"
	"github.com/camaeel/vault-autounseal-operator/pkg/config"
	vaultapi "github.com/hashicorp/vault/api"
	corev1 "k8s.io/api/core/v1"
)

func GetVaultClient(cfg *config.Config, pod corev1.Pod) (*vaultapi.Client, error) {
	defaultCfg := vaultapi.DefaultConfig() // modify for more granular configuration
	defaultCfg.Address = fmt.Sprintf("%s://%s.%s:%d", "https", pod.Name, cfg.ServiceDomain, cfg.ServicePort)

	tlsConfig := vaultapi.TLSConfig{
		CACert: cfg.CaCert,
	}
	err := defaultCfg.ConfigureTLS(&tlsConfig)
	if err != nil {
		return nil, err
	}

	client, err := vaultapi.NewClient(defaultCfg)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize vault client: %w", err)
	}

	return client, err
}
