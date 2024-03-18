// Copyright 2023 Cisco Systems, Inc. and its affiliates
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package jaegertracing

import (
    "context"
    "fmt"
    "github.com/cisco-open/go-lanai/pkg/bootstrap"
	"github.com/cisco-open/go-lanai/pkg/log"
	"github.com/cisco-open/go-lanai/pkg/tracing"
	"github.com/opentracing/opentracing-go"
    "github.com/uber/jaeger-client-go"
    "io"
)

var logger = log.New("Tracing")

func NewDefaultTracer() (opentracing.Tracer, io.Closer) {
	return newTracer("lanai", jaeger.NewConstSampler(false), jaeger.NewNullReporter())
}

func NewTracer(ctx *bootstrap.ApplicationContext, jp *tracing.JaegerProperties, sp *tracing.SamplerProperties) (opentracing.Tracer, io.Closer) {
	name := ctx.Name()
	sampler := newSampler(ctx, sp)
	reporter := newReporter(ctx, jp, sp)
	return newTracer(name, sampler, reporter)
}

// newTracer we use B3 single header compatible format, this is compatible with Spring-Sleuth powered services
// See https://github.com/openzipkin/b3-propagation#single-header
// See https://github.com/jaegertracing/jaeger-client-go/blob/master/zipkin/README.md#NewZipkinB3HTTPHeaderPropagator
func newTracer(serviceName string, sampler jaeger.Sampler, reporter jaeger.Reporter,) (opentracing.Tracer, io.Closer) {
	b3HttpPropagator := NewZipkinB3Propagator()
	b3SingleHeaderPropagator := NewZipkinB3Propagator(SingleHeader())
	zipkinOpts := []jaeger.TracerOption {
		jaeger.TracerOptions.Injector(opentracing.HTTPHeaders, b3HttpPropagator),
		jaeger.TracerOptions.Injector(opentracing.TextMap, b3SingleHeaderPropagator),
		jaeger.TracerOptions.Extractor(opentracing.HTTPHeaders, b3HttpPropagator),
		jaeger.TracerOptions.Extractor(opentracing.TextMap, b3SingleHeaderPropagator),
		// Zipkin shares span ID between client and server spans; it must be enabled via the following option.
		jaeger.TracerOptions.ZipkinSharedRPCSpan(true),
	}
	return jaeger.NewTracer(serviceName, sampler, reporter, zipkinOpts...)
}

func newSampler(ctx context.Context, sp *tracing.SamplerProperties) jaeger.Sampler {
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

func newReporter(ctx context.Context, jp *tracing.JaegerProperties, sp *tracing.SamplerProperties) jaeger.Reporter {
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
