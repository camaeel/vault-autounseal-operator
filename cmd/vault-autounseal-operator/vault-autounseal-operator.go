package main

import (
	"context"
	"flag"
	"time"

	"github.com/camaeel/vault-autounseal-operator/pkg/config"
	"github.com/camaeel/vault-autounseal-operator/pkg/utils/logging"
	vaultUnsealOperator "github.com/camaeel/vault-autounseal-operator/pkg/vault-autounseal-operator"
)

func main() {
	cfg := &config.Config{}

	flag.StringVar(&cfg.PodSelector, "pod-selector", "app.kubernetes.io/instance=vault,app.kubernetes.io/name=vault", "Selector for finding vault's pods")
	flag.StringVar(&cfg.StatefulsetSelector, "statefulset-selector", "", "Selector for finding vault's statefulsets. If empty, then pod-selector is used")
	flag.StringVar(&cfg.Namespace, "namespace", "vault", "Namespace running vault")

	flag.StringVar(&cfg.ServiceDomain, "service-domain", "vault-internal.vault.svc.cluster.local", "DNS Name for accessing vault. In HA mode should be set to vault headles service providing all pod endpoints.")
	flag.StringVar(&cfg.ServiceScheme, "service-scheme", "https", "Vault service scheme. Valid values: http, https")
	flag.IntVar(&cfg.ServicePort, "service-port", 8200, "Vaul service port")
	flag.StringVar(&cfg.VaultCaCertPath, "ca-cert-path", "", "Path to vault CA certificate")
	flag.BoolVar(&cfg.TlsSkipVerify, "tls-skip-verify", false, "Skip vault TLS certificate verification, default: false")

	flag.IntVar(&cfg.UnlockShares, "unlock-shares", 1, "Number of unlock shares")
	flag.IntVar(&cfg.UnlockThreshold, "unlock-threshold", 1, "Number of unlock shares threshold")
	flag.StringVar(&cfg.VaultRootTokenSecret, "vault-root-token-secret", "vault-root-token", "Vault root token secret name")
	flag.StringVar(&cfg.VaultUnlockKeysSecret, "vault-unlock-keys-secret", "vault-unlock-keys", "Vault unlock keys secret name")
	// kubeconfig := flag.String("kubeconfig", "", "Overwrite kubeconfig path")

	flag.StringVar(&cfg.LeaseName, "leader-election-lease-name", "vault-autounseal-leader", "Name of the lease object for leader election")
	flag.StringVar(&cfg.LeaseNamespace, "leader-election-lease-namespace", "", "Name of the namespace with lease object for leader election. If empty use the same namespace as the application is running in")
	flag.DurationVar(&cfg.InformerResync, "resync-period", 60*time.Second, "Reconcilation loop frequency")

	flag.StringVar(&cfg.LogFormat, "log-format", "json", "Log format. Allowed values are: text, json. Default is json. ")
	flag.StringVar(&cfg.LogLevel, "log-level", "info", "Log level. Allowed values are: debug, info, warn, error. Default is info.")

	flag.StringVar(&cfg.HandlerTimeout, "timeout", "30s", "Operator handler timeout. Default is 30s")

	flag.Parse()
	err := cfg.Initialize()
	if err != nil {
		panic(err)
	}
	err = cfg.Validate()
	if err != nil {
		panic(err)
	}

	ctx := context.TODO()

	err = logging.SetupLogging(cfg)
	if err != nil {
		panic(err)
	}

	vaultUnsealOperator.Exec(ctx, cfg)
}
