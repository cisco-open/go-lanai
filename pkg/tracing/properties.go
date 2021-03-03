package tracing

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"github.com/pkg/errors"
)

const (
	TracingPropertiesPrefix = "tracing"
)

type TracingProperties struct {
	Enabled bool              `json:"enabled"`
	Jaeger  JaegerProperties  `json:"jaeger"`
	Zipkin  ZipkinProperties  `json:"zipkin"`
	Sampler SamplerProperties `json:"sampler"`
}

type JaegerProperties struct {
	Enabled bool   `json:"enabled"`
	Host    string `json:"host"`
	Port    int    `json:"port"`
}

type ZipkinProperties struct {
	Enabled bool `json:"enabled"`
}

type SamplerProperties struct {
	Enabled     bool    `json:"enabled"`
	RateLimit   float64 `json:"limit-per-second"`
	Probability float64 `json:"probability"`
	LowestRate  float64 `json:"lowest-per-second"`
}

//NewSessionProperties create a SessionProperties with default values
func NewTracingProperties() *TracingProperties {
	return &TracingProperties{
		Enabled: true,
		Jaeger: JaegerProperties{
			Enabled: true,
		},
		Zipkin: ZipkinProperties{},
		Sampler: SamplerProperties{
			Enabled:   false,
			RateLimit: 10.0,
		},
	}
}

//BindManagementProperties create and bind SessionProperties, with a optional prefix
func BindTracingProperties(ctx *bootstrap.ApplicationContext) TracingProperties {
	props := NewTracingProperties()
	if err := ctx.Config().Bind(props, TracingPropertiesPrefix); err != nil {
		panic(errors.Wrap(err, "failed to bind TracingProperties"))
	}
	return *props
}
