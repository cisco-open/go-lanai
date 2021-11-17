package data

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/order"
	"gorm.io/gorm"
)

// ErrorTranslator redefines web.ErrorTranslator and order.Ordered
// having this redefinition is to break dependency between data and web package
type ErrorTranslator interface {
	order.Ordered
	Translate(ctx context.Context, err error) error
}

type DbCreator interface {
	CreateDatabaseIfNotExist(ctx context.Context, db *gorm.DB) error
}
