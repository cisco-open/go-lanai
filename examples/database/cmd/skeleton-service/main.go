package main

import (
	serviceinit "cto-github.cisco.com/NFV-BU/go-lanai/examples/skeleton-service/pkg/init"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"go.uber.org/fx"
	"time"
)

func init() {
	// initialize modules
	serviceinit.Use()

	//gin.SetMode(gin.ReleaseMode)
}

func main() {
	// bootstrapping
	bootstrap.NewAppCmd(
		"skeleton-service",
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
