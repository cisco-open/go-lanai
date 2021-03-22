package data

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/order"
	"go.uber.org/fx"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
	"time"
)

type gormInitDI struct {
	fx.In
	Dialector gorm.Dialector
	Properties DataProperties
	Configurers [] GormConfigurer `group:"gorm_config"`
}

type GormConfigurer interface {
	Configure(config *gorm.Config)
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

	config := gorm.Config{
		Logger: newGormLogger(level, slow),
	}

	// gave configurer an chance
	order.SortStable(di.Configurers, order.OrderedFirstCompare)
	for _, c := range di.Configurers {
		c.Configure(&config)
	}

	db, e := gorm.Open(di.Dialector, &config)
	if e != nil {
		panic(e)
	}
	return db
}
