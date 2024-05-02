package config

import (
	"fmt"
	"strings"
	"time"

	"k8s.io/client-go/kubernetes"
)

type Config struct {
	LogFormat           string
	Namespace           string
	PodSelectorMap      map[string]string
	StsSelectorMap      map[string]string
	PodSelector         string
	StatefulsetSelector string
	LeaseName           string
	LeaseNamespace      string
	K8sClient           kubernetes.Interface
	InformerResync      time.Duration
	CaCertPath          string
}

func (cfg *Config) Validate() error {
	if cfg.LogFormat != "text" && cfg.LogFormat != "json" {
		return fmt.Errorf("wrong log format %s. Allowed values are: json, text", cfg.LogFormat)
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
