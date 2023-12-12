package acm

import (
	awsclient "cto-github.cisco.com/NFV-BU/go-lanai/pkg/aws"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"github.com/aws/aws-sdk-go-v2/service/acm"
	"go.uber.org/fx"
)

var logger = log.New("Aws")

var Module = &bootstrap.Module{
	Name:       "ACM",
	Precedence: bootstrap.AwsPrecedence,
	Options: []fx.Option{
		fx.Provide(NewClientFactory),
		fx.Provide(NewDefaultClient),
		fx.Invoke(registerHealth),
	},
}

// Use func, does nothing. Allow service to include this module in main()
func Use() {
	bootstrap.Register(Module)
	bootstrap.Register(awsclient.Module)
}

func NewDefaultClient(ctx *bootstrap.ApplicationContext, factory ClientFactory) (*acm.Client, error) {
	return factory.New(ctx)
}
