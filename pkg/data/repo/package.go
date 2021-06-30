package repo

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"go.uber.org/fx"
)

//var logger = log.New("DB.Repo")

var Module = &bootstrap.Module{
	Name: "DB Repo",
	Precedence: bootstrap.DatabasePrecedence,
	Options: []fx.Option{
		fx.Provide(newGormFactory),
		fx.Provide(newGormApi),
	},
}
