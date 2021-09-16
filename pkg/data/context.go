package data

import (
	"context"
)

type DbCreator interface {
	CreateDatabaseIfNotExist(ctx context.Context) error
}