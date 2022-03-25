package cockroach

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/data"
	"fmt"
	"gorm.io/gorm"
)

type GormDbCreator struct {
	dbUser string
	dbName string
}

func NewGormDbCreator(properties CockroachProperties) data.DbCreator {
	return &GormDbCreator{
		dbUser: properties.Username,
		dbName: properties.Database,
	}
}

func (g *GormDbCreator) CreateDatabaseIfNotExist(ctx context.Context, db *gorm.DB) error {
	if g.dbUser != "root" {
		logger.WithContext(ctx).Info("db user is not a privileged account, skipped db creation.")
		return nil
	}
	result := db.WithContext(ctx).Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", g.dbName))
	return result.Error

}