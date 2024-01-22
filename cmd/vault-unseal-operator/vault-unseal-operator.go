package main

import (
	"context"
	"flag"

	"github.com/camaeel/vault-unseal-operator/pkg/config"
	"github.com/camaeel/vault-unseal-operator/pkg/utils/logger"
	vaultUnsealOperator "github.com/camaeel/vault-unseal-operator/pkg/vault-unseal-operator"
)

func main() {
	cfg := &config.Config{}

	flag.StringVar(&cfg.LogFormat, "log-format", "json", "Log format. Allowed values are: text, json. Default is json. ")
	// flag.StringVar(&cfg.ServiceDomain, "service-domain", "vault-internal.vault.svc.cluster.local", "DNS Name for accessing vault. In HA mode should be set to vault headles service providing all pod endpoints.")
	// flag.StringVar(&cfg.ServiceScheme, "service-scheme", "https", "Vaul service scheme. Valid values: http, https")
	// flag.IntVar(&cfg.ServicePort, "service-port", 8200, "Vaul service port")
	// flag.IntVar(&cfg.UnlockShares, "unlock-shares", 3, "Number of unlock shares")
	// flag.IntVar(&cfg.UnlockThreshold, "unlock-threshold", 2, "Number of unlock shares threshold")
	// flag.StringVar(&cfg.VaultRootTokenSecret, "vault-root-token-secret", "vault-root-token", "Vault root token secret name")
	// flag.StringVar(&cfg.VaultUnlockKeysSecret, "vault-unlock-keys-secret", "vault-unlock-keys", "Vault unlock keys secret name")
	flag.StringVar(&cfg.PodSelector, "pod-selector", "app.kubernetes.io/instance=vault,app.kubernetes.io/name=vault", "Selector for finding vault's pods")
	flag.StringVar(&cfg.StatefulsetSelector, "statefulset-selector", "", "Selector for finding vault's statefulsets. If empty, then pod-selector is used")
	flag.StringVar(&cfg.Namespace, "namespace", "vault", "Namespace running vault")
	// kubeconfig := flag.String("kubeconfig", "", "Overwrite kubeconfig path")

	flag.Parse()
	err := cfg.Validate()
	if err != nil {
		panic(err)
	}

	ctx := context.TODO()

	logger.SetupLogging(cfg)
	vaultUnsealOperator.Exec(ctx)
}
