package vault

import (
	"context"
	"fmt"
	"github.com/camaeel/vault-autounseal-operator/pkg/config"
	vault "github.com/hashicorp/vault/api"
	"log/slog"
)

type Node struct {
	Client *vault.Client
}

func (n *Node) Initialize(cfg *config.Config, ctx context.Context) ([]string, string, error) {
	req := vault.InitRequest{
		SecretShares:    cfg.UnlockShares,
		SecretThreshold: cfg.UnlockThreshold,
	}
	resp, err := n.Client.Sys().InitWithContext(ctx, &req)
	if err != nil {
		return nil, "", err
	}
	keys := resp.Keys
	rootToken := resp.RootToken
	return keys, rootToken, nil
}

func (n *Node) Unseal(logger *slog.Logger, ctx context.Context, keys [][]byte) error {
	for i := range keys {
		logger.Info(fmt.Sprintf("Unsealing vault node with key number %d", i))
		_, err := n.Client.Sys().UnsealWithContext(ctx, string(keys[i]))
		if err != nil {
			logger.Error(fmt.Sprintf("Failed to unseal vault node with key number %d", i))
			return err
		}
	}

	return nil
}

func (n *Node) GetSealStatus(ctx context.Context) (sealed bool, initialized bool, err error) {
	resp, err := n.Client.Sys().SealStatusWithContext(ctx)
	if err != nil {
		return
	}
	sealed = resp.Sealed
	initialized = resp.Initialized

	return
}
