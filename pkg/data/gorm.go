package data

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"go.uber.org/fx"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
	"time"
)

type gormInitDI struct {
	fx.In
	Dialector gorm.Dialector
	Properties DataProperties
}

func NewGorm(di gormInitDI) *gorm.DB {
	level := gormlogger.Warn
	switch di.Properties.Logging.Level {
	case log.LevelOff:
		level = gormlogger.Silent
	case log.LevelDebug, log.LevelInfo:
		level = gormlogger.Info
	case log.LevelWarn:
		level = gormlogger.Warn
	case log.LevelError:
		level = gormlogger.Error
	}

	slow := time.Duration(di.Properties.Logging.SlowThreshold)
	if slow == 0 {
		slow = 15 * time.Second
	}

	db, e := gorm.Open(di.Dialector, &gorm.Config{
		Logger: newGormLogger(level, slow),
	})
	if e != nil {
		panic(e)
	}
	return db
}
