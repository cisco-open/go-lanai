package tracing

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"fmt"
	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	"io"
)

func NewDefaultTracer() (opentracing.Tracer, io.Closer) {
	return jaeger.NewTracer("lanai", jaeger.NewConstSampler(false), jaeger.NewNullReporter())
}

func NewJaegerTracer(ctx *bootstrap.ApplicationContext, jp *JaegerProperties, sp *SamplerProperties) (opentracing.Tracer, io.Closer) {
	name := ctx.Name()
	sampler := newSampler(ctx, sp)
	reporter := newReporter(ctx, jp, sp)
	return jaeger.NewTracer(name, sampler, reporter)
}

func newSampler(ctx context.Context, sp *SamplerProperties) jaeger.Sampler {
	if !sp.Enabled {
		return jaeger.NewConstSampler(false)
	}

	if sp.LowestRate > 0 && sp.Probability > 0 && sp.Probability <= 1.0 {
		sampler, e := jaeger.NewGuaranteedThroughputProbabilisticSampler(sp.LowestRate, sp.Probability)
		if e == nil {
			logger.WithContext(ctx).
				Infof("Use GuaranteedThroughputProbabilisticSampler with lowest rate %.3f/s and probability %%%2.1f",
				sp.LowestRate, sp.Probability * 100)
			return sampler
		}
	}

	if sp.Probability > 0 && sp.Probability <= 1.0 {
		sampler, e := jaeger.NewProbabilisticSampler(sp.Probability)
		if e == nil {
			logger.WithContext(ctx).
				Infof("Use ProbabilisticSampler with lprobability %%%2.1f", sp.Probability * 100)
			return sampler
		}
	}

	if sp.RateLimit > 0 {
		sampler := jaeger.NewRateLimitingSampler(sp.RateLimit)
		logger.WithContext(ctx).
			Infof("Use RateLimitingSampler with rate limit %.3f/s", sp.RateLimit)
		return sampler
	}

	logger.WithContext(ctx).Warnf("both rate limit and probability are not valid, tracing sampling is disabled")
	return jaeger.NewConstSampler(false)
}

func newReporter(ctx context.Context, jp *JaegerProperties, sp *SamplerProperties) jaeger.Reporter {
	if !sp.Enabled || jp.Host == "" || jp.Port == 0 {
		return jaeger.NewNullReporter()
	}

	hostPort := fmt.Sprintf("%s:%d", jp.Host, jp.Port)
	transport, e := jaeger.NewUDPTransport(hostPort, 0)
	if e != nil {
		panic(fmt.Sprintf("unable to estabilish connection to Jaeger server at %s", hostPort))
	}

	return jaeger.NewRemoteReporter(transport)
}
