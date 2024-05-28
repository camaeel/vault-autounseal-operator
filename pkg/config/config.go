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

	Namespace           string
	PodSelector         string
	StatefulsetSelector string

	K8sClient             kubernetes.Interface
	InformerResync        time.Duration
	VaultCaCertPath       string
	VaultCaCert           string
	TlsSkipVerify         bool
	UnlockShares          int
	UnlockThreshold       int
	ServiceDomain         string
	ServicePort           int
	ServiceScheme         string
	VaultRootTokenSecret  string
	VaultUnlockKeysSecret string
}

func (cfg *Config) Initialize() error {
	if cfg.VaultCaCertPath != "" {
		cacert, err := os.ReadFile(cfg.VaultCaCertPath)
		if err != nil {
			return err
		}
		cfg.VaultCaCert = string(cacert)
	}
	return nil
}

func (cfg *Config) Validate() error {
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
