package acm

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/actuator/health"
	"github.com/aws/aws-sdk-go-v2/service/acm"
	"go.uber.org/fx"
)

type regDI struct {
	fx.In
	HealthRegistrar health.Registrar `optional:"true"`
	AcmClient       *acm.Client
}

func registerHealth(di regDI) {
	if di.HealthRegistrar == nil {
		return
	}
	di.HealthRegistrar.MustRegister(&HealthIndicator{
		AcmClient: di.AcmClient,
	})
}

// HealthIndicator monitor ACM client status
type HealthIndicator struct {
	AcmClient *acm.Client
}

func (i *HealthIndicator) Name() string {
	return "aws.acm"
}

func (i *HealthIndicator) Health(ctx context.Context, options health.Options) health.Health {
	input := &acm.GetAccountConfigurationInput{}
	if _, e := i.AcmClient.GetAccountConfiguration(ctx, input); e != nil {
		logger.WithContext(ctx).Warnf("AWS ACM connection not available or identity invalid: %v", e)
		return health.NewDetailedHealth(health.StatusUnknown, "AWS ACM connection not available or identity invalid", nil)
	} else {
		return health.NewDetailedHealth(health.StatusUp, "aws connect succeeded", nil)
	}
}
