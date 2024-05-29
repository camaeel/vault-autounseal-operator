package vault

import (
	"fmt"
	"github.com/camaeel/vault-autounseal-operator/pkg/config"
	vaultapi "github.com/hashicorp/vault/api"
	corev1 "k8s.io/api/core/v1"
	"log/slog"
)

func GetVaultClient(cfg *config.Config, pod *corev1.Pod) (*vaultapi.Client, error) {
	defaultCfg := vaultapi.DefaultConfig() // modify for more granular configuration
	defaultCfg.Address = fmt.Sprintf("%s://%s.%s:%d", cfg.ServiceScheme, pod.Name, cfg.ServiceDomain, cfg.ServicePort)
	//defaultCfg.Address = fmt.Sprintf("%s://%s:%d", cfg.ServiceScheme, "127.0.0.1", cfg.ServicePort)
	defaultCfg.Timeout = cfg.HandlerTimeoutDuration
	defaultCfg.MaxRetries = 0

	tlsConfig := vaultapi.TLSConfig{
		CACert:        cfg.VaultCaCert,
		TLSServerName: fmt.Sprintf("%s.%s", pod.Name, cfg.ServiceDomain),
		//Insecure:      true,
		Insecure: cfg.TlsSkipVerify,
	}
	err := defaultCfg.ConfigureTLS(&tlsConfig)
	if err != nil {
		return nil, err
	}

	client, err := vaultapi.NewClient(defaultCfg)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize vault client: %w", err)
	}

	slog.Debug("Created vault client", "pod", pod.Name, "address", defaultCfg.Address)

	return client, err
}
