package data

import (
	"go.uber.org/fx"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
	"time"
)

type gormInitDI struct {
	fx.In
	Dialector gorm.Dialector
}

func NewGorm(di gormInitDI) *gorm.DB {
	db, e := gorm.Open(di.Dialector, &gorm.Config{
		// TODO value from properties
		Logger: newGormLogger(gormlogger.Info, 1 * time.Second),
	})
	if e != nil {
		panic(e)
	}
	return db
}
