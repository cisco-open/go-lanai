package kafka

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"go.uber.org/fx"
)

var logger = log.New("Kafka")

var Module = &bootstrap.Module{
	Precedence: bootstrap.KafkaPrecedence,
	Options: []fx.Option{
		fx.Provide(BindKafkaProperties),
		fx.Provide(NewSaramaProducerFactory),
		fx.Invoke(registerHealth),
	},
}

const (
	FxGroup = "kafka"
)

// Use Allow service to include this module in main()
func Use() {
	bootstrap.Register(Module)
}