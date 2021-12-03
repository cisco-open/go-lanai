package apilist

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/actuator"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"go.uber.org/fx"
	"io/fs"
	"os"
)

var logger = log.New("ACTR.APIList")

var staticFS = []fs.FS{os.DirFS(".")}

var Module = &bootstrap.Module{
	Name:       "actuator-apilist",
	Precedence: actuator.MinActuatorPrecedence,
	Options: []fx.Option{
		fx.Provide(BindProperties),
	},
}

func init() {
	bootstrap.Register(Module)
}

func StaticFS(fs ...fs.FS) {
	if len(fs) != 0 {
		staticFS = fs
	}
}

type regDI struct {
	fx.In
	Registrar     *actuator.Registrar
	MgtProperties actuator.ManagementProperties
	Properties    Properties
}

func Register(di regDI) {
	ep := newEndpoint(di)
	di.Registrar.MustRegister(ep)
}
