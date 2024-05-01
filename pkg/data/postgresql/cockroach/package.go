package cockroach

import (
	"github.com/cisco-open/go-lanai/pkg/bootstrap"
	"github.com/cisco-open/go-lanai/pkg/data/postgresql"
	"github.com/cisco-open/go-lanai/pkg/log"
	"go.uber.org/fx"
)

var logger = log.New("cockroach")

var Module = &bootstrap.Module{
	Name:       "cockroach",
	Precedence: bootstrap.DatabasePrecedence,
	Options: []fx.Option{
		fx.Provide(newAnnotatedGormDbCreator()),
	},
}

func Use() {
	bootstrap.Register(postgresql.Module)
	bootstrap.Register(Module)
}
