package vaultUnsealOperator

import (
	"context"
	"fmt"

	"github.com/camaeel/vault-unseal-operator/pkg/utils/logger"
)

func Exec(ctx context.Context) {
	logger.Logger().Info(fmt.Sprintf("Staring now %s asd", "asd"))
}
