package acmcerts

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/certs"
	certsource "cto-github.cisco.com/NFV-BU/go-lanai/pkg/certs/source"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/service/acm/acmiface"
	"go.uber.org/fx"
)

var logger = log.New("Certs.ACM")

const (
	sourceType = certs.SourceACM
)

type factoryDI struct {
	fx.In
	AppCtx    *bootstrap.ApplicationContext
	Props     certs.Properties `optional:"true"`
	AcmClient acmiface.ACMAPI  `optional:"true"`
}

func FxProvider() fx.Annotated {
	return fx.Annotated{
		Group: certs.FxGroup,
		Target: func(di factoryDI) (certs.SourceFactory, error) {
			if di.AcmClient == nil {
				logger.Warnf(`Vault Certificates source is not supported. Tips: Do not forget to initialize vault client.`)
				return nil, nil
			}

			var rawDefaults json.RawMessage
			if di.Props.Sources != nil {
				rawDefaults, _ = di.Props.Sources[sourceType]
			}
			factory, e := certsource.NewFactory[SourceProperties](sourceType, rawDefaults, func(props SourceProperties) certs.Source {
				return NewAcmProvider(di.AppCtx, di.AcmClient, props)
			})
			if e != nil {
				return nil, fmt.Errorf(`unable to register certificate source type [%s]: %v`, sourceType, e)
			}
			return factory, nil
		},
	}
}
