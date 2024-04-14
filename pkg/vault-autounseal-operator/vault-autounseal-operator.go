package vaultAutounsealOperator

import (
	"context"
	"fmt"

	"github.com/camaeel/vault-autounseal-operator/pkg/utils/logger"
)

func Exec(ctx context.Context) {
	logger.Logger().Info(fmt.Sprintf("Staring now %s asd", "asd"))
}
