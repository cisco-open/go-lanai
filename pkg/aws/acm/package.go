package acm

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"github.com/aws/aws-sdk-go/service/acm/acmiface"
	"go.uber.org/fx"
)

var logger = log.New("Aws")

type AcmClient struct {
	Client acmiface.ACMAPI
}

var Module = &bootstrap.Module{
	Name:       "ACM",
	Precedence: bootstrap.AcmPresedence,
	Options: []fx.Option{
		fx.Provide(BindAwsProperties),
		fx.Provide(NewAwsAcmFactory),
		fx.Provide(newDefaultClient),
		fx.Invoke(registerHealth),
	},
}

// Use func, does nothing. Allow service to include this module in main()
func Use() {
	bootstrap.Register(Module)
}

func newDefaultClient(ctx *bootstrap.ApplicationContext, f AwsAcmFactory) AcmClient {
	c, e := f.New(ctx)

	if e != nil {
		panic(e)
	}
	return AcmClient{c}
}
