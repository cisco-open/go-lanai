package repo

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"go.uber.org/fx"
)

//var logger = log.New("DB.Repo")

var globalFactory Factory

var Module = &bootstrap.Module{
	Name: "DB Repo",
	Precedence: bootstrap.DatabasePrecedence,
	Options: []fx.Option{
		fx.Provide(newGormFactory),
		fx.Provide(newGormApi),
		fx.Invoke(initialize),
	},
}

func initialize(factory Factory) {
	globalFactory = factory
	defaultUtils = newGormUtils(factory.(*GormFactory))
}
