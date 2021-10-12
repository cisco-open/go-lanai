package cockroach

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/data"
	"fmt"
	"gorm.io/gorm"
)

type GormDbCreator struct {
	dbName string
}

func NewGormDbCreator(properties CockroachProperties) data.DbCreator {
	return &GormDbCreator{
		dbName: properties.Database,
	}
}

func (g *GormDbCreator) CreateDatabaseIfNotExist(ctx context.Context, db *gorm.DB) error {
	result := db.WithContext(ctx).Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", g.dbName))
	return result.Error
}