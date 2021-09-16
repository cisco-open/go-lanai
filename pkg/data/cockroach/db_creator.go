package cockroach

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/data"
	"fmt"
	"gorm.io/gorm"
)

type GormDbCreator struct {
	properties CockroachProperties
	db *gorm.DB
}

func NewGormDbCreator(properties CockroachProperties, db *gorm.DB) data.DbCreator {
	return &GormDbCreator{
		properties: properties,
		db: db,
	}
}

func (g *GormDbCreator) CreateDatabaseIfNotExist(ctx context.Context) error {
	result := g.db.WithContext(ctx).Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", g.properties.Database))
	return result.Error
}