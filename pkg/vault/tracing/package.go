package vaulttracing

import (
	"github.com/cisco-open/go-lanai/pkg/bootstrap"
	"github.com/cisco-open/go-lanai/pkg/vault"
	"github.com/opentracing/opentracing-go"
	"go.uber.org/fx"
)

var Module = &bootstrap.Module{
	Name:       "vault-tracing",
	Precedence: bootstrap.TracingPrecedence,
	PriorityOptions: []fx.Option{
		fx.Invoke(initialize),
	},
}

type tracerDI struct {
	fx.In
	AppContext   *bootstrap.ApplicationContext
	Tracer       opentracing.Tracer  `optional:"true"`
	VaultClient  *vault.Client       `optional:"true"`
	// we could include security configurations, customizations here
}

func initialize(di tracerDI) {
	// vault instrumentation
	if di.Tracer != nil && di.VaultClient != nil {
		hook := NewHook(di.Tracer)
		di.VaultClient.AddHooks(di.AppContext, hook)
	}
}