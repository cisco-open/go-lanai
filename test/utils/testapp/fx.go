package testapp

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	test "cto-github.cisco.com/NFV-BU/go-lanai/test/utils"
	"go.uber.org/fx"
	"testing"
	"time"
)

type testModuleCtxKey struct{}

var ctxKeyTestModule = testModuleCtxKey{}

type testFxContext struct {
	context.Context
	module *bootstrap.Module
}

func (c *testFxContext) Value(key interface{}) interface{} {
	switch {
	case key == ctxKeyTestModule:
		return c.module
	}
	return c.Context.Value(key)
}

// WithDI populate given di targets by using fx.Populate
// all targets need to be pointer to struct, otherwise the test fails
// See fx.Populate for more information
func WithDI(diTargets ...interface{}) test.Options {
	return WithFxOptions(fx.Populate(diTargets...))
}

// WithModules register given modules to test app
func WithModules(modules ...*bootstrap.Module) test.Options {
	return test.Setup(func(ctx context.Context, t *testing.T) (context.Context, error) {
		for _, m := range modules {
			bootstrap.Register(m)
		}
		return ctx, nil
	})
}

func WithTimeout(timeout time.Duration) test.Options {
	return WithFxOptions(fx.StartTimeout(timeout))
}

// WithFxOptions register given fx.Option to test app as regular steps
// see bootstrap.Module
func WithFxOptions(opts ...fx.Option) test.Options {
	return test.Setup(func(ctx context.Context, t *testing.T) (context.Context, error) {
		ret, fxOpts := withTestModule(ctx)
		fxOpts.Options = append(fxOpts.Options, opts...)
		return ret, nil
	})
}

func WithFxPriorityOptions(opts ...fx.Option) test.Options {
	return test.Setup(func(ctx context.Context, t *testing.T) (context.Context, error) {
		ret, fxOpts := withTestModule(ctx)
		fxOpts.PriorityOptions = append(fxOpts.PriorityOptions, opts...)
		return ret, nil
	})
}

func withTestModule(ctx context.Context) (context.Context, *bootstrap.Module) {
	ret := ctx
	m, ok := ctx.Value(ctxKeyTestModule).(*bootstrap.Module)
	if !ok || m == nil {
		m = &bootstrap.Module{
			Name:            "test",
			Precedence:      bootstrap.LowestPrecedence,
			PriorityOptions: []fx.Option{},
			Options:         []fx.Option{},
		}
		ret = &testFxContext{
			Context: ctx,
			module:  m,
		}
	}
	return ret, m
}