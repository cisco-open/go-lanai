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

package redis_test

import (
    "context"
    "embed"
    "github.com/cisco-open/go-lanai/pkg/actuator"
    "github.com/cisco-open/go-lanai/pkg/actuator/health"
    "github.com/cisco-open/go-lanai/pkg/bootstrap"
    certsinit "github.com/cisco-open/go-lanai/pkg/certs/init"
    "github.com/cisco-open/go-lanai/pkg/redis"
    "github.com/cisco-open/go-lanai/test"
    "github.com/cisco-open/go-lanai/test/apptest"
    "github.com/cisco-open/go-lanai/test/embedded"
    goRedis "github.com/go-redis/redis/v8"
    "github.com/onsi/gomega"
    . "github.com/onsi/gomega"
    "go.uber.org/fx"
    "testing"
)

/*************************
	Test Setup
 *************************/

//go:embed testdata/*
var TLSConfigFS embed.FS

func ProvideTestHooks() *TestHook {
	return &TestHook{}
}

func RegisterTestHooks(appCtx *bootstrap.ApplicationContext, factory redis.ClientFactory, hook *TestHook) {
	factory.AddHooks(appCtx, hook)
}

/*************************
	Tests
 *************************/

type TestDI struct {
	fx.In
	DefaultClient   redis.Client
	ClientFactory   redis.ClientFactory
	HealthIndicator health.Indicator `optional:"true"`
	TestHook        *TestHook
}

func TestRedisConnectivity(t *testing.T) {
	di := &TestDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		embedded.WithRedis(),
		apptest.WithModules(redis.Module, actuator.Module, health.Module),
		apptest.WithDI(di),
		apptest.WithFxOptions(
			fx.Provide(ProvideTestHooks),
			fx.Invoke(RegisterTestHooks),
		),
		test.GomegaSubTest(SubTestDefaultClient(di), "DefaultClient"),
		test.GomegaSubTest(SubTestClientFactory(di), "TestClientFactory"),
		test.GomegaSubTest(SubTestHealthIndicator(di), "TestHealthIndicator"),
	)
}

func TestRedisTLSConnectivity(t *testing.T) {
	di := &TestDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		embedded.WithRedis(embedded.EnableTLS(func(src *embedded.TLSCerts) {
			src.FS = TLSConfigFS
		})),
		apptest.WithModules(redis.Module, certsinit.Module),
		apptest.WithDI(di),
		apptest.WithFxOptions(
			fx.Provide(ProvideTestHooks),
			fx.Invoke(RegisterTestHooks),
		),
		apptest.WithProperties("redis.tls.enabled: true", "redis.tls.certs.type: file"),
		test.GomegaSubTest(SubTestDefaultClient(di), "DefaultClient"),
		test.GomegaSubTest(SubTestClientFactory(di), "TestClientFactory"),
	)
}

/*************************
	Sub Tests
 *************************/

func SubTestDefaultClient(di *TestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		// use factory
		AssertClient(ctx, g, di.DefaultClient, di.TestHook)
	}
}

func SubTestClientFactory(di *TestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		// use factory
		client, e := di.ClientFactory.New(ctx, func(opt *redis.ClientOption) {
			opt.DbIndex = 5
		})
		g.Expect(e).To(Succeed(), "injected client factory shouldn't return error")
		AssertClient(ctx, g, client, di.TestHook)
	}
}

func SubTestHealthIndicator(di *TestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		h := di.HealthIndicator.Health(ctx, health.Options{
			ShowDetails:    true,
			ShowComponents: true,
		})
		g.Expect(h).To(Not(BeNil()), "Health status should not be nil")
		g.Expect(h.Status()).To(BeEquivalentTo(health.StatusUp), "Health status should be UP")
	}
}

/*************************
	Helpers
 *************************/

func AssertClient(ctx context.Context, g *gomega.WithT, client redis.Client, hook *TestHook) {
	before := len(hook.Commands)
	// ping
	cmd := client.Ping(ctx)
	g.Expect(cmd).To(Not(BeNil()), "redis ping shouldn't return nil")
	g.Expect(cmd.Err()).To(Succeed(), "redis ping shouldn't return error")
	g.Expect(hook.Commands).To(HaveLen(before+1), "redis hook should be invoked")
}

type TestHook struct {
	Commands   []goRedis.Cmder
	Successful []goRedis.Cmder
	Failed     []goRedis.Cmder
}

func (h *TestHook) BeforeProcess(ctx context.Context, cmd goRedis.Cmder) (context.Context, error) {
	h.Commands = append(h.Commands, cmd)
	return ctx, nil
}

func (h *TestHook) AfterProcess(_ context.Context, cmd goRedis.Cmder) error {
	if cmd.Err() != nil {
		h.Failed = append(h.Failed, cmd)
	} else {
		h.Successful = append(h.Successful, cmd)
	}
	return nil
}

func (h *TestHook) BeforeProcessPipeline(ctx context.Context, cmds []goRedis.Cmder) (context.Context, error) {
	for _, cmd := range cmds {
		ctx, _ = h.BeforeProcess(ctx, cmd)
	}
	return ctx, nil
}

func (h *TestHook) AfterProcessPipeline(ctx context.Context, cmds []goRedis.Cmder) error {
	for _, cmd := range cmds {
		_ = h.AfterProcess(ctx, cmd)
	}
	return nil
}

func (h *TestHook) WithClientOption(_ *goRedis.UniversalOptions) goRedis.Hook {
	return h
}
