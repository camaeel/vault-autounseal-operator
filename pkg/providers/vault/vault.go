package vault

import (
	"context"
	"fmt"
	"github.com/camaeel/vault-autounseal-operator/pkg/config"
	corev1 "k8s.io/api/core/v1"

	vault "github.com/hashicorp/vault/api"
)

func newVaultClient(cfg *config.Config, pod *corev1.Pod) (*vault.Client, error) {

	defaultConfig := vault.DefaultConfig() // modify for more granular configuration
	defaultConfig.Timeout = cfg.VaultTimeoutDuration
	//defaultCfg.MaxRetries = 0?
	if val, found := cfg.PodAddressesMap[pod.Name]; found {
		defaultConfig.Address = val
	} else {
		defaultConfig.Address = fmt.Sprintf("%s://%s.%s:%d", cfg.ServiceScheme, pod.Name, cfg.ServiceDomain, cfg.ServicePort)
	}

	tlsConfig := vault.TLSConfig{
		CACert:        cfg.VaultCaCert,
		Insecure:      cfg.TlsSkipVerify,
		TLSServerName: fmt.Sprintf("%s.%s", pod.Name, cfg.ServiceDomain),
	}
	err := defaultConfig.ConfigureTLS(&tlsConfig)
	if err != nil {
		return nil, err
	}

	client, err := vault.NewClient(defaultConfig)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize vault client: %w", err)
	}

	return client, nil
}

func GetVaultClusterNode(ctx context.Context, cfg *config.Config, pod *corev1.Pod) (Node, error) {
	var err error
	var node Node
	node.Client, err = newVaultClient(cfg, pod)

	return node, err
}
