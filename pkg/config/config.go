package config

import (
	"fmt"
	"strings"
	"time"

	"k8s.io/client-go/kubernetes"
)

type Config struct {
	LogFormat string

	LeaseName      string
	LeaseNamespace string

	Namespace           string
	PodSelectorMap      map[string]string
	StsSelectorMap      map[string]string
	PodSelector         string
	StatefulsetSelector string

	K8sClient             kubernetes.Interface
	InformerResync        time.Duration
	CaCertPath            string
	TlsSkipVerify         bool
	UnlockShares          int
	UnlockThreshold       int
	ServiceDomain         string
	ServicePort           int
	ServiceScheme         string
	VaultRootTokenSecret  string
	VaultUnlockKeysSecret string
}

func (cfg *Config) Validate() error {
	if cfg.LogFormat != "text" && cfg.LogFormat != "json" {
		return fmt.Errorf("wrong log format %s. Allowed values are: json, text", cfg.LogFormat)
	}
	if cfg.ServiceScheme != "http" && cfg.ServiceScheme != "https" {
		return fmt.Errorf("wrong service scheme %s. Allowed values are: http, https", cfg.ServiceScheme)
	}

	podSelectorMap, err := parseMap(cfg.PodSelector)
	if err != nil {
		return fmt.Errorf("wrong pod selector format %s, due to %v", cfg.PodSelector, err)
	}
	cfg.PodSelectorMap = podSelectorMap

	cfg.StsSelectorMap, err = parseMap(cfg.StatefulsetSelector)
	if err != nil {
		return fmt.Errorf("wrong statefulset selector format %s, due to %v", cfg.StatefulsetSelector, err)

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
