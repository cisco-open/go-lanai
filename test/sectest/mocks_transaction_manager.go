package testscope

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/data/tx"
	txMocks "cto-github.cisco.com/NFV-BU/go-lanai/pkg/data/tx/mocks"
)

func WithMockedTransactionManager(ctx context.Context) *txMocks.TxManager{
	txManager := new(txMocks.TxManager)
	tx.NewTxManager(txManager)
	return txManager
}