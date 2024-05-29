package podhandler

import (
	"context"
	"github.com/camaeel/vault-autounseal-operator/pkg/config"
	vaulthttp "github.com/hashicorp/vault/http"
	"github.com/hashicorp/vault/vault"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"log/slog"
	"strconv"
	"testing"
)

func TestGetPodHandlerFunctions(t *testing.T) {
	cfg := config.Config{
		Namespace: "vault",
	}
	ret := GetPodHandlerFunctions(&cfg, context.TODO(), nil)
	assert.NotNil(t, ret)
}

func TestIsInitialized_True(t *testing.T) {
	pod := corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-pod",
			Labels: map[string]string{
				"vault-initialized": "true",
			},
		},
	}

	res := isInitialized(&pod)
	assert.True(t, res)
}

func TestIsInitialized_False(t *testing.T) {
	pod := corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-pod",
			Labels: map[string]string{
				"vault-initialized": "false",
			},
		},
	}

	res := isInitialized(&pod)
	assert.False(t, res)
}

func TestIsInitialized_MissingAnnotation(t *testing.T) {
	pod := corev1.Pod{}

	res := isInitialized(&pod)
	assert.False(t, res)
}

func TestIsInitialized_InvalidAnnotationValue(t *testing.T) {
	pod := corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-pod",
			Labels: map[string]string{
				"vault-initialized": "invalid",
			},
		},
	}

	res := isInitialized(&pod)
	assert.False(t, res)
}

func TestIsSealed_True(t *testing.T) {
	pod := corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-pod",
			Labels: map[string]string{
				"vault-sealed": "true",
			},
		},
	}

	res := isSealed(&pod)
	assert.True(t, res)
}

func TestIsSealed_False(t *testing.T) {
	pod := corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-pod",
			Labels: map[string]string{
				"vault-sealed": "false",
			},
		},
	}

	res := isSealed(&pod)
	assert.False(t, res)
}

func TestIsSealed_MissingAnnotation(t *testing.T) {
	pod := corev1.Pod{}

	res := isSealed(&pod)
	assert.False(t, res)
}

func TestIsSealed_InvalidAnnotationValue(t *testing.T) {
	pod := corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-pod",
			Labels: map[string]string{
				"vault-sealed": "invalid",
			},
		},
	}

	res := isSealed(&pod)
	assert.False(t, res)
}

func TestIsLeader_True(t *testing.T) {
	pod := corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-pod",
			Labels: map[string]string{
				"vault-active": "true",
			},
		},
	}

	res := isLeader(&pod)
	assert.True(t, res)
}

func TestIsLeader_False(t *testing.T) {
	pod := corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-pod",
			Labels: map[string]string{
				"vault-active": "false",
			},
		},
	}

	res := isLeader(&pod)
	assert.False(t, res)
}

func TestIsLeader_MissingAnnotation(t *testing.T) {
	pod := corev1.Pod{}

	res := isLeader(&pod)
	assert.False(t, res)
}

func TestIsLeader_InvalidAnnotationValue(t *testing.T) {
	pod := corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-pod",
			Annotations: map[string]string{
				"vault-active": "invalid",
			},
		},
	}

	res := isLeader(&pod)
	assert.False(t, res)
}

func TestInitialize(t *testing.T) {
	ctx := context.TODO()
	cfg := config.Config{
		UnlockShares:    3,
		UnlockThreshold: 3,
	}
	// mock Vault server
	coreConfig := vault.CoreConfig{}
	opts := vault.TestClusterOptions{
		SkipInit:    true,
		NumCores:    3,
		HandlerFunc: vaulthttp.Handler,
	}
	testVault := vault.NewTestCluster(t, &coreConfig, &opts)
	vaultClient := testVault.Cores[0].Client

	initData, err := initialize(&cfg, ctx, vaultClient)
	assert.NoError(t, err)
	assert.NotNil(t, initData)
	assert.Len(t, initData.Keys, 3)

	sealStatus, err := vaultClient.Sys().SealStatus()
	assert.NoError(t, err)
	assert.NotNil(t, sealStatus)
	assert.True(t, sealStatus.Initialized)
	assert.True(t, sealStatus.Sealed)
}

func TestUnseal(t *testing.T) {
	ctx := context.TODO()
	// mock Vault server
	coreConfig := vault.CoreConfig{}
	opts := vault.TestClusterOptions{
		SkipInit:    true,
		NumCores:    3,
		HandlerFunc: vaulthttp.Handler,
	}
	testVault := vault.NewTestCluster(t, &coreConfig, &opts)
	vaultClient := testVault.Cores[0].Client

	cfg := config.Config{
		UnlockShares:    3,
		UnlockThreshold: 3,
	}

	//need to initialize first
	initData, err := initialize(&cfg, ctx, vaultClient)

	sealStatusBefore, err := vaultClient.Sys().SealStatus()
	assert.NoError(t, err)
	assert.NotNil(t, sealStatusBefore)
	assert.True(t, sealStatusBefore.Initialized)
	assert.True(t, sealStatusBefore.Sealed)

	unsealKeys := map[string][]byte{}
	for k, v := range initData.Keys {
		unsealKeys[strconv.Itoa(k)] = []byte(v)
	}

	err = unseal(slog.Default(), ctx, vaultClient, unsealKeys)
	assert.NoError(t, err)

	sealStatus, err := vaultClient.Sys().SealStatus()
	assert.NoError(t, err)
	assert.NotNil(t, sealStatus)
	assert.True(t, sealStatus.Initialized)
	assert.False(t, sealStatus.Sealed)
}
