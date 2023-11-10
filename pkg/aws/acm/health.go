package acm

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/actuator/health"
	"github.com/aws/aws-sdk-go/service/acm"
	"go.uber.org/fx"
)

type regDI struct {
	fx.In
	HealthRegistrar health.Registrar `optional:"true"`
	AcmClient       AcmClient
}

func registerHealth(di regDI) {
	if di.HealthRegistrar == nil {
		return
	}
	di.HealthRegistrar.MustRegister(&AcmHealthIndicator{
		AcmClient: di.AcmClient,
	})
}

// AwsHealthIndicator
type AcmHealthIndicator struct {
	AcmClient AcmClient
}

func (i *AcmHealthIndicator) Name() string {
	return "aws.acm"
}

func (i *AcmHealthIndicator) Health(c context.Context, options health.Options) health.Health {
	input := &acm.GetAccountConfigurationInput{}
	if _, e := i.AcmClient.Client.GetAccountConfigurationWithContext(c, input); e != nil {
		logger.WithContext(c).Errorf("AWS ACM connection not available or identity invalid: %v", e)
		return health.NewDetailedHealth(health.StatusUnknown, "AWS ACM connection not available or identity invalid", nil)
	} else {
		return health.NewDetailedHealth(health.StatusUp, "aws connect succeeded", nil)
	}
}
