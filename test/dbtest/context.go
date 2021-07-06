package dbtest

import (
	"context"
	"gorm.io/gorm"
)

type mockedTxContext struct {
	context.Context
}

func (c mockedTxContext) Parent() context.Context {
	return c.Context
}

type mockedGormContext struct {
	mockedTxContext
	db *gorm.DB
}

func (c mockedGormContext) DB() *gorm.DB {
	return c.db
}

