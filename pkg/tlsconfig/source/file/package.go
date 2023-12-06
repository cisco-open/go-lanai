package filecerts

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/tlsconfig"
	certsource "cto-github.cisco.com/NFV-BU/go-lanai/pkg/tlsconfig/source"
	"fmt"
	"go.uber.org/fx"
)

const (
	sourceType = tlsconfig.SourceFile
)

type factoryDI struct {
	fx.In
	Props tlsconfig.Properties
}

func FxProvider() fx.Annotated {
	return fx.Annotated{
		Group: tlsconfig.FxGroup,
		Target: func(di factoryDI) (tlsconfig.SourceFactory, error) {
			rawDefaults, _ := di.Props.Sources[sourceType]
			factory, e := certsource.NewFactory[SourceProperties](sourceType, rawDefaults, NewFileProvider)
			if e != nil {
				return nil, fmt.Errorf(`unable to register certificate source type [%s]: %v`, sourceType, e)
			}
			return factory, nil
		},
	}
}
