package acmcerts

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/tlsconfig"
	certsource "cto-github.cisco.com/NFV-BU/go-lanai/pkg/tlsconfig/source"
	"fmt"
	"github.com/aws/aws-sdk-go/service/acm/acmiface"
	"go.uber.org/fx"
)

var logger = log.New("Certs.ACM")

const (
	sourceType = tlsconfig.SourceACM
)

type factoryDI struct {
	fx.In
	Props     tlsconfig.Properties
	AcmClient acmiface.ACMAPI `optional:"true"`
}

func FxProvider() fx.Annotated {
	return fx.Annotated{
		Group: tlsconfig.FxGroup,
		Target: func(di factoryDI) (tlsconfig.SourceFactory, error) {
			if di.AcmClient == nil {
				logger.Warnf(`Vault Certificates source is not supported. Tips: Do not forget to initialize vault client.`)
				return nil, nil
			}

			rawDefaults, _ := di.Props.Sources[sourceType]
			factory, e := certsource.NewFactory[SourceProperties](sourceType, rawDefaults, func(props SourceProperties) tlsconfig.Source {
				return NewAcmProvider(di.AcmClient, props)
			})
			if e != nil {
				return nil, fmt.Errorf(`unable to register certificate source type [%s]: %v`, sourceType, e)
			}
			return factory, nil
		},
	}
}
