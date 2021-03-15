package data

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"go.uber.org/fx"
)

var logger = log.New("Data")

var Module = &bootstrap.Module{
	Name: "DB",
	Precedence: bootstrap.DatabasePrecedence,
	Options: []fx.Option{
		fx.Provide(BindDataProperties),
		fx.Invoke(registerHealth),
	},
}


/**************************
	Provider
***************************/


/**************************
	Initialize
***************************/




