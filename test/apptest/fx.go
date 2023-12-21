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

package apptest

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"go.uber.org/fx"
	"testing"
	"time"
)

// WithDI populate given di targets by using fx.Populate
// all targets need to be pointer to struct, otherwise the test fails
// See fx.Populate for more information
func WithDI(diTargets ...interface{}) test.Options {
	return WithFxOptions(fx.Populate(diTargets...))
}

// WithModules register given modules to test app
func WithModules(modules ...*bootstrap.Module) test.Options {
	return test.Setup(func(ctx context.Context, t *testing.T) (context.Context, error) {
		ret, tb := withTestModule(ctx)
		for _, m := range modules {
			tb.Register(m)
		}
		return ret, nil
	})
}

// WithTimeout specify expected test timeout to prevent blocking test process permanently
func WithTimeout(timeout time.Duration) test.Options {
	return WithFxOptions(fx.StartTimeout(timeout))
}

// WithFxOptions register given fx.Option to test app as regular steps
// see bootstrap.Module
func WithFxOptions(opts ...fx.Option) test.Options {
	return test.Setup(func(ctx context.Context, t *testing.T) (context.Context, error) {
		ret, tb := withTestModule(ctx)
		tb.AddOptions(opts...)
		return ret, nil
	})
}

// WithFxPriorityOptions register given fx.Option to test app as priority steps, before any other modules
// see bootstrap.Module
func WithFxPriorityOptions(opts ...fx.Option) test.Options {
	return test.Setup(func(ctx context.Context, t *testing.T) (context.Context, error) {
		ret, tb := withTestModule(ctx)
		tb.AddOptions()
		tb.AppPriorityOptions = append(tb.AppPriorityOptions, opts...)
		return ret, nil
	})
}

func withTestModule(ctx context.Context) (context.Context, *testBootstrapper) {
	ret := ctx
	tb, ok := ctx.Value(ctxKeyTestBootstrapper).(*testBootstrapper)
	if !ok || tb == nil {
		tb = &testBootstrapper{
			Bootstrapper: *bootstrap.NewBootstrapper(),
		}
		ret = &testFxContext{
			Context: ctx,
			tb:      tb,
		}
	}
	return ret, tb
}

/*********************
	Test FX Context
 *********************/
type testBootstrapperCtxKey struct{}

var ctxKeyTestBootstrapper = testBootstrapperCtxKey{}

type testFxContext struct {
	context.Context
	tb *testBootstrapper
}

func (c *testFxContext) Value(key interface{}) interface{} {
	switch {
	case key == ctxKeyTestBootstrapper:
		return c.tb
	}
	return c.Context.Value(key)
}
