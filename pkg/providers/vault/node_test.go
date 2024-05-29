package vault

import (
	"context"
	"github.com/camaeel/vault-autounseal-operator/pkg/config"
	"github.com/stretchr/testify/assert"
	"log/slog"
	"testing"
)

func TestGetSealStatus(t *testing.T) {
	testVault := GetVault(t, true, 1)

	n := Node{
		Client: testVault[0].Client,
	}
	sealed, initialized, err := n.GetSealStatus(context.TODO())
	assert.NoError(t, err)
	assert.True(t, initialized)
	assert.False(t, sealed)
}

func TestInitialize(t *testing.T) {
	testVault := GetVault(t, false, 1)

	n := Node{
		Client: testVault[0].Client,
	}
	cfg := config.Config{
		UnlockThreshold: 3,
		UnlockShares:    5,
	}

	unsealKeys, rootToken, err := n.Initialize(&cfg, context.TODO())
	assert.NoError(t, err)
	assert.Len(t, unsealKeys, 5)

	assert.GreaterOrEqual(t, len(rootToken), 1)

	sealed, initialized, err := n.GetSealStatus(context.TODO())
	assert.NoError(t, err)
	assert.True(t, initialized)
	assert.True(t, sealed)
}

func TestUnseal(t *testing.T) {
	testVault := GetVault(t, false, 1)

	n := Node{
		Client: testVault[0].Client,
	}
	cfg := config.Config{
		UnlockThreshold: 3,
		UnlockShares:    5,
	}

	unsealKeys, _, err := n.Initialize(&cfg, context.TODO())
	assert.NoError(t, err)

	unsealKeyBytes := [][]byte{}
	for i := range unsealKeys {
		unsealKeyBytes = append(unsealKeyBytes, []byte(unsealKeys[i]))
	}

	err = n.Unseal(slog.Default(), context.TODO(), unsealKeyBytes)
	assert.NoError(t, err)

	sealed, initialized, err := n.GetSealStatus(context.TODO())
	assert.NoError(t, err)
	assert.True(t, initialized)
	assert.False(t, sealed)
}
