package postgresql

import (
	"context"
	"github.com/cisco-open/go-lanai/pkg/data"
	"go.uber.org/fx"
	"gorm.io/gorm"
)

const (
	DBCreatorPostgresOrder = iota
)

type NoOpDbCreator struct{}

func (g NoOpDbCreator) Order() int {
	return DBCreatorPostgresOrder
}

func (g NoOpDbCreator) CreateDatabaseIfNotExist(ctx context.Context, db *gorm.DB) error {
	// postgres can't connect to database if it doesn't exist, nothing to do here
	return nil
}

func NewGormDbCreator() data.DbCreator {
	return &NoOpDbCreator{}
}

func newAnnotatedGormDbCreator() fx.Annotated {
	return fx.Annotated{
		Group:  data.GormConfigurerGroup,
		Target: NewGormDbCreator,
	}
}
