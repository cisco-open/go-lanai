package data

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/actuator/health"
	"go.uber.org/fx"
	"gorm.io/gorm"
)

type regDI struct {
	fx.In
	HealthRegistrar health.Registrar `optional:"true"`
	GormDB *gorm.DB `optional:"true"`
}

func registerHealth(di regDI) {
	if di.HealthRegistrar == nil || di.GormDB == nil {
		return
	}
	di.HealthRegistrar.MustRegister(&DbHealthIndicator{
		db: di.GormDB,
	})
}

// DbHealthIndicator
// Note: we currently only support one database
type DbHealthIndicator struct {
	db *gorm.DB
}

func (i *DbHealthIndicator) Name() string {
	return "database"
}

func (i *DbHealthIndicator) Health(c context.Context, options health.Options) health.Health {
	if sqldb, e := i.db.DB(); e != nil {
		return health.NewDetailedHealth(health.StatusUnknown, "database ping is not available", nil)
	} else {
		if e := sqldb.Ping(); e != nil {
			return health.NewDetailedHealth(health.StatusDown, "database ping failed", nil)
		} else {
			return health.NewDetailedHealth(health.StatusUp, "database ping succeeded", nil)
		}
	}
}


