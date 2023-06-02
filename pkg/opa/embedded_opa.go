package opa

import (
	"bytes"
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"encoding/json"
	"fmt"
	"github.com/ghodss/yaml"
	"github.com/open-policy-agent/opa/sdk"
	sdktest "github.com/open-policy-agent/opa/sdk/test"
	"go.uber.org/fx"
	"io"
)

var embeddedOPA *sdk.OPA

type EmbeddedOPAReadyCH <-chan struct{}

type EmbeddedOPAOut struct {
	fx.Out
	OPA   *sdk.OPA
	Ready EmbeddedOPAReadyCH
}

func EmbeddedOPA() *sdk.OPA {
	return embeddedOPA
}

func ProvideEmbeddedOPA(appCtx *bootstrap.ApplicationContext, bundleServer *sdktest.Server) (EmbeddedOPAOut, error) {
	cfg, e := LoadConfig(appCtx, bundleServer)
	if e != nil {
		return EmbeddedOPAOut{}, fmt.Errorf("unable to load OPA config: %v", e)
	}
	opaLog := NewOPALogger(logger.WithContext(appCtx), log.LevelInfo)
	ready := make(chan struct{}, 1)
	opa, e := sdk.New(appCtx, sdk.Options{
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

// LoadConfig
// TODO POC only
func LoadConfig(appCtx *bootstrap.ApplicationContext, bundleServer *sdktest.Server) (io.Reader, error){
	baseCfg, e := ConfigFS.Open("opa-config.yml")
	if e != nil {
		return nil, e
	}
	baseYml, e := io.ReadAll(baseCfg)
	if e != nil {
		return nil, e
	}
	baseJson, e := yaml.YAMLToJSON(baseYml)
	if e != nil {
		return nil, e
	}
	var cfg map[string]interface{}
	if e := json.Unmarshal(baseJson, &cfg); e != nil {
		return nil, e
	}

	pocJson := fmt.Sprintf(`{
		"services": {
			"poc": {
				"url": %q
			}
		},
		"bundles": {
			"api": {
				"resource": "/bundles/api.tar.gz"
			}
		}
	}`, bundleServer.URL())

	if e := json.Unmarshal([]byte(pocJson), &cfg); e != nil {
		return nil, e
	}
	cfgJson, e := json.Marshal(cfg)
	if e != nil {
		return nil, e
	}
	return bytes.NewReader(cfgJson), nil
}
