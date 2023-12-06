package acm

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"github.com/aws/aws-sdk-go/service/acm/acmiface"
	"go.uber.org/fx"
)

var logger = log.New("Aws")

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

func newDefaultClient(ctx *bootstrap.ApplicationContext, f AwsAcmFactory) (acmiface.ACMAPI, error) {
	return f.New(ctx)
}
