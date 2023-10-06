package main

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	serviceinit "cto-github.cisco.com/NFV-BU/test-service/pkg/init"
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
		"testservice",
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
