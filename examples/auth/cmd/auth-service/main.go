package main

import (
	serviceinit "cto-github.cisco.com/NFV-BU/go-lanai/examples/auth-service/pkg/init"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"go.uber.org/fx"
	"time"
)

func init() {
	serviceinit.Use()
}

func main() {
	// bootstrapping
	bootstrap.NewAppCmd(
		"authservice",
		[]fx.Option{
			// Some priority fx.Provide() and fx.Invoke()
		},
		[]fx.Option{
			fx.StartTimeout(60 * time.Second),
			// fx.Provide(),
		},
	)
	bootstrap.Execute()
}
