package acmcerts

import (
	awsclient "cto-github.cisco.com/NFV-BU/go-lanai/pkg/aws"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/certs"
	certsource "cto-github.cisco.com/NFV-BU/go-lanai/pkg/certs/source"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/service/acm"
	"go.uber.org/fx"
)

var logger = log.New("Certs.ACM")

const (
	sourceType = certs.SourceACM
)

var Module = &bootstrap.Module{
	Name:       "certs-acm",
	Precedence: bootstrap.TlsConfigPrecedence,
	Options: []fx.Option{
		fx.Provide(FxProvider()),
	},
}

func Use() {
	bootstrap.Register(Module)
}

type factoryDI struct {
	fx.In
	AppCtx          *bootstrap.ApplicationContext
	Props           certs.Properties        `optional:"true"`
	AcmClient       *acm.Client             `optional:"true"`
	AwsConfigLoader awsclient.ConfigLoader `optional:"true"`
}

func FxProvider() fx.Annotated {
	return fx.Annotated{
		Group: certs.FxGroup,
		Target: func(di factoryDI) (certs.SourceFactory, error) {
			var client *acm.Client
			switch {
			case di.AcmClient == nil && di.AwsConfigLoader == nil:
				logger.Warnf(`AWS/ACM certificates source is not supported. Tips: Do not forget to initialize ACM client or AWS config loader.`)
				return nil, nil
			case di.AcmClient != nil:
				client = di.AcmClient
			default:
				cfg, e := di.AwsConfigLoader.Load(di.AppCtx)
				if e != nil {
					return nil, fmt.Errorf(`unable to initialize AWS/ACM certificate source: %v`, e)
				}
				client = acm.NewFromConfig(cfg)
			}

			var rawDefaults json.RawMessage
			if di.Props.Sources != nil {
				rawDefaults, _ = di.Props.Sources[sourceType]
			}
			factory, e := certsource.NewFactory[SourceProperties](sourceType, rawDefaults, func(props SourceProperties) certs.Source {
				return NewAcmProvider(di.AppCtx, client, props)
			})
			if e != nil {
				return nil, fmt.Errorf(`unable to register certificate source type [%s]: %v`, sourceType, e)
			}
			return factory, nil
		},
	}
}
