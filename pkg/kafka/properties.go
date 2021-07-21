package kafka

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"github.com/pkg/errors"
)

const (
	ConfigKafkaPrefix = "kafka"
)

type KafkaProperties struct {
	Brokers string `json:"brokers"`
	Net Net `json:"net"`
}

type Net struct {
	Sasl SASL `json:"sasl"`
}

type SASL struct {
	// Whether or not to use SASL authentication when connecting to the broker
	// (defaults to false).
	Enable bool `json:"enabled"`
	// Whether or not to send the Kafka SASL handshake first if enabled
	// (defaults to true). You should only set this to false if you're using
	// a non-Kafka SASL proxy.
	Handshake bool `json:"handshake"`
	//username and password for SASL/PLAIN authentication
	User     string `json:"user"`
	Password string `josn:"password"`
}

func BindKafkaProperties(ctx *bootstrap.ApplicationContext) KafkaProperties {
	props := KafkaProperties{
		Net : Net{
			Sasl: SASL{
				Enable: false,
				Handshake: true,
			},
		},
	}
	if err := ctx.Config().Bind(&props, ConfigKafkaPrefix); err != nil {
		panic(errors.Wrap(err, "failed to bind redis.RedisProperties"))
	}
	return props
}
