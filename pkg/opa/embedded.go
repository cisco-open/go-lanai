package opa

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"fmt"
	"github.com/open-policy-agent/opa/sdk"
	"go.uber.org/fx"
)

var embeddedOPA *sdk.OPA

type EmbeddedOPAReadyCH <-chan struct{}

type EmbeddedOPAOut struct {
	fx.Out
	OPA   *sdk.OPA
	Ready EmbeddedOPAReadyCH
}

type EmbeddedOPADI struct {
	fx.In
	AppCtx      *bootstrap.ApplicationContext
	Properties  Properties
	Customizers []ConfigCustomizer `group:"opa"`
}

func EmbeddedOPA() *sdk.OPA {
	return embeddedOPA
}

func ProvideEmbeddedOPA(di EmbeddedOPADI) (EmbeddedOPAOut, error) {
	cfg, e := LoadConfig(di.AppCtx, di.Properties, di.Customizers...)
	if e != nil {
		return EmbeddedOPAOut{}, fmt.Errorf("unable to load OPA Config: %v", e)
	}
	opaLog := NewOPALogger(logger.WithContext(di.AppCtx), log.LevelInfo)
	ready := make(chan struct{}, 1)
	opa, e := sdk.New(di.AppCtx, sdk.Options{
		ID:            `Embedded-OPA`,
		Config:        cfg,
		Logger:        opaLog,
		ConsoleLogger: opaLog,
		Ready:         ready,
		Plugins:       nil,
	})
	if e != nil {
		close(ready)
		return EmbeddedOPAOut{}, fmt.Errorf("error when create embedded OPA: %v", e)
	}
	return EmbeddedOPAOut{
		OPA:   opa,
		Ready: ready,
	}, nil
}

func InitializeEmbeddedOPA(lc fx.Lifecycle, opa *sdk.OPA, ready EmbeddedOPAReadyCH) {
	embeddedOPA = opa
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			select {
			case <-ready:
				logger.WithContext(ctx).Infof("Embedded OPA is Ready")
				return nil
			case <-ctx.Done():
				return fmt.Errorf("embedded OPA is failed to start")
			}
		},
		OnStop: func(ctx context.Context) error {
			opa.Stop(ctx)
			return nil
		},
	})
}
