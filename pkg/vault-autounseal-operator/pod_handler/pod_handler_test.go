package podhandler

import (
	"context"
	"github.com/camaeel/vault-autounseal-operator/pkg/config"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetPodHandlerFunctions(t *testing.T) {
	cfg := config.Config{
		Namespace: "vault",
	}
	ret := GetPodHandlerFunctions(&cfg, context.TODO(), nil)
	assert.NotNil(t, ret)
}
