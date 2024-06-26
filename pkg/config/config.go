package config

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"k8s.io/client-go/kubernetes"
)

type Config struct {
	LogFormat string
	LogLevel  string

	LeaseName      string
	LeaseNamespace string

	Namespace   string
	PodSelector string
	//StatefulsetSelector string

	K8sClient            kubernetes.Interface
	InformerResync       time.Duration
	VaultCaCertPath      string
	VaultCaCert          string
	TlsSkipVerify        bool
	VaultTimeout         string
	VaultTimeoutDuration time.Duration

	PodAddresses    string
	PodAddressesMap map[string]string

	UnlockShares          int
	UnlockThreshold       int
	ServiceDomain         string
	ServicePort           int
	ServiceScheme         string
	VaultRootTokenSecret  string
	VaultUnlockKeysSecret string

	HandlerTimeout         string
	HandlerTimeoutDuration time.Duration

	Port int
}

func (cfg *Config) InitializeAndValidate() error {
	var err error
	if cfg.VaultCaCertPath != "" {
		_, err := os.ReadFile(cfg.VaultCaCertPath)
		if err != nil {
			return fmt.Errorf("error reading CA cert file: %v", err)
		}
	}

	if cfg.PodAddresses != "" {
		cfg.PodAddressesMap, err = parseMap(cfg.PodAddresses)
		if err != nil {
			return fmt.Errorf("error parsing pod addresses: %v", err)
		}
	}

	if cfg.LogFormat != "text" && cfg.LogFormat != "json" {
		return fmt.Errorf("wrong log format %s. Allowed values are: json, text", cfg.LogFormat)
	}
	if cfg.ServiceScheme != "http" && cfg.ServiceScheme != "https" {
		return fmt.Errorf("wrong service scheme %s. Allowed values are: http, https", cfg.ServiceScheme)
	}
	if cfg.LogLevel != strings.ToLower(slog.LevelDebug.String()) &&
		cfg.LogLevel != strings.ToLower(slog.LevelInfo.String()) &&
		cfg.LogLevel != strings.ToLower(slog.LevelWarn.String()) &&
		cfg.LogLevel != strings.ToLower(slog.LevelError.String()) {
		return fmt.Errorf("wrong log level %s. Allowed values are: debug, info, warn, error", cfg.LogLevel)
	}

	cfg.HandlerTimeoutDuration, err = time.ParseDuration(cfg.HandlerTimeout)
	if err != nil {
		return fmt.Errorf("wrong duration for handler timeout: %v", err)
	}

	cfg.VaultTimeoutDuration, err = time.ParseDuration(cfg.VaultTimeout)
	if err != nil {
		return fmt.Errorf("wrong duration for vault timeout: %v", err)
	}

	if cfg.Port < 1 || cfg.Port > 65535 {
		return fmt.Errorf("invalid port")
	}

	return nil
}

func parseMap(str string) (map[string]string, error) {
	ret := map[string]string{}

	if str == "" {
		return ret, nil
	}

	strs := strings.Split(str, ",")
	for i := range strs {
		keyvalues := strings.Split(strs[i], "=")
		if len(keyvalues) != 2 {
			return map[string]string{}, fmt.Errorf("Wrong number of = sings between colons in %s", strs[i])
		}
		ret[keyvalues[0]] = keyvalues[1]
	}
	return ret, nil

}
