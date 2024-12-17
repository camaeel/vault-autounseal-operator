package vault

import (
	"testing"

	vaulthttp "github.com/hashicorp/vault/http"
	"github.com/hashicorp/vault/vault"
)

func GetVault(t *testing.T, initialize bool, cores int) []*vault.TestClusterCore {
	t.Helper()

	coreConfig := vault.CoreConfig{}
	opts := vault.TestClusterOptions{
		SkipInit:    !initialize,
		NumCores:    cores,
		HandlerFunc: vaulthttp.Handler,
	}
	testVault := vault.NewTestCluster(t, &coreConfig, &opts)
	return testVault.Cores

}
