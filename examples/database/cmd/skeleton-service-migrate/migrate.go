package main

import (
	"github.com/cisco-open/go-lanai/examples/skeleton-service/pkg/migrate"
	"github.com/cisco-open/go-lanai/pkg/bootstrap"
	"go.uber.org/fx"
	"time"
)

func init() {
	migrate.Use() // This line initialize the data migration implementations
}

func main() {
	// bootstrapping
	bootstrap.NewAppCmd(
		"migrate",
		nil,
		[]fx.Option{fx.StartTimeout(525600 * time.Minute)}, //We can't have this timeout. Adding long timeout
	)

	bootstrap.Execute()
}
