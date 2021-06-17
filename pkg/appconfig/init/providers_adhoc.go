package appconfig

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/appconfig"
	"go.uber.org/fx"
)

type adhocBootstrapDI struct {
	fx.In
	Providers []appconfig.Provider `group:"bootstrap-config"`
}

func newBootstrapAdHocProviderGroup(di adhocBootstrapDI) bootstrapProvidersOut {
	providers := make([]appconfig.Provider, 0)
	for _, p := range di.Providers {
		if p == nil {
			continue
		}
		if reorder, ok := p.(appconfig.ProviderReorderer); ok {
			reorder.Reorder(bootstrapAdHocPrecedence)
		}
		providers =  append(providers, p)
	}
	return bootstrapProvidersOut {
		ProviderGroup: appconfig.NewStaticProviderGroup(bootstrapAdHocPrecedence, providers...),
	}
}

type adhocApplicationDI struct {
	fx.In
	Providers []appconfig.Provider `group:"application-config"`
}

func newApplicationAdHocProviderGroup(di adhocApplicationDI) appConfigProvidersOut {
	providers := make([]appconfig.Provider, 0)
	for _, p := range di.Providers {
		if p == nil {
			continue
		}
		if reorder, ok := p.(appconfig.ProviderReorderer); ok {
			reorder.Reorder(applicationAdHocPrecedence)
		}
		providers =  append(providers, p)
	}
	return appConfigProvidersOut {
		ProviderGroup: appconfig.NewStaticProviderGroup(applicationAdHocPrecedence, providers...),
	}
}
