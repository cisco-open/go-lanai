package aws

import (
    "cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
    "github.com/aws/aws-sdk-go-v2/config"
    "go.uber.org/fx"
)

const FxGroup = `aws`

var Module = &bootstrap.Module{
	Name:       "AWS",
	Precedence: bootstrap.AwsPrecedence,
	Options: []fx.Option{
		fx.Provide(BindAwsProperties, ProvideConfigLoader),
	},
}

func FxCustomizerProvider(constructor interface{}) fx.Annotated {
	return fx.Annotated{
		Group:  FxGroup,
		Target: constructor,
	}
}

type CfgLoaderDI struct {
	fx.In
	Properties  Properties
	Customizers []config.LoadOptionsFunc `group:"aws"`
}

func ProvideConfigLoader(di CfgLoaderDI) ConfigLoader {
	return NewConfigLoader(di.Properties, di.Customizers...)
}
