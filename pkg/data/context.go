package data

import (
	"context"
	"gorm.io/gorm"
)

type DbCreator interface {
	CreateDatabaseIfNotExist(ctx context.Context, db *gorm.DB) error
}