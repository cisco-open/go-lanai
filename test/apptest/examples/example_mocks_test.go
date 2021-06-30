package examples

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/data/tx"
	"database/sql"
	"gorm.io/gorm"
)


type mockedService struct {}

func (t *mockedService) DummyMethod(_ context.Context) error {
	return nil
}

func NewMockedService() DummyService {
	return &mockedService{}
}

type noopTxManager struct {}

func provideNoopTxManager() tx.TxManager {
	return noopTxManager{}
}

func (m noopTxManager) Transaction(_ context.Context, _ tx.TxFunc, _ ...*sql.TxOptions) error {
	return nil
}

func (m noopTxManager) WithDB(_ *gorm.DB) tx.GormTxManager {
	return m
}


