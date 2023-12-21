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
	appconfig "cto-github.cisco.com/NFV-BU/go-lanai/pkg/appconfig/init"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"embed"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
	"reflect"
	"testing"
	"time"
)

//go:embed test-defaults.yml
var TestDefaultConfigFS embed.FS

//go:embed test-bootstrap.yml
var TestBootstrapConfigFS embed.FS

//go:embed test-application.yml
var TestApplicationConfigFS embed.FS

// testBootstrapper holds all configuration to bootstrap a fs-enabled test
type testBootstrapper struct {
	bootstrap.Bootstrapper
	AppPriorityOptions []fx.Option
	AppOptions         []fx.Option
}

// Bootstrap is an entrypoint test.Options that indicates all sub tests should be run within the scope of
// an slim version of bootstrap.App
func Bootstrap() test.Options {
	return test.WithInternalRunner(NewFxTestRunner())
}

// NewFxTestRunner is internal use only, exported for cross-package reference
func NewFxTestRunner() test.InternalRunner {
	return func(ctx context.Context, t *test.T) {
		// run setup hooks
		ctx = testSetup(ctx, t.T, t.TestHooks)
		defer testTeardown(ctx, t.T, t.TestHooks)

		// register test module's options without register the module directly
		// Note:
		//		we want to support repeated bootstrap but the bootstrap package doesn't support
		// 		module refresh (caused by singleton pattern).
		// Note 4.3:
		//		Now with help of bootstrap.ExecuteContainedApp(), we are able repeatedly bootstrap a self-contained
		// 		application.
		tb, ok := ctx.Value(ctxKeyTestBootstrapper).(*testBootstrapper)
		if !ok || tb == nil {
			t.Fatalf("Failed to start test %s due to invalid test bootstrap configuration", t.Name())
			return
		}

		// default modules and context
		tb.Register(appconfig.Module)
		tb.AddInitialAppContextOptions(mergeInitContext(ctx))

		// prepare bootstrap fx options
		priority := append([]fx.Option{
			fx.Supply(t),
			appconfig.FxEmbeddedDefaults(TestDefaultConfigFS),
			appconfig.FxEmbeddedBootstrapAdHoc(TestBootstrapConfigFS),
			appconfig.FxEmbeddedApplicationAdHoc(TestApplicationConfigFS),
		}, tb.AppPriorityOptions...)
		regular := append([]fx.Option{}, tb.AppOptions...)

		// bootstrapping
		//nolint:contextcheck // context is not passed on because the bootstrap process is not cancellable. This is a limitation
		bootstrap.NewAppCmd("testapp", priority, regular,
			func(cmd *cobra.Command) {
				cmd.Use = "testapp"
				cmd.Args = nil
			},
		)
		tb.EnableCliRunnerMode(newTestCliRunner)
		bootstrap.ExecuteContainedApp(ctx, &tb.Bootstrapper)
	}
}

func newTestCliRunner(t *test.T) bootstrap.CliRunner {
	return func(ctx context.Context) error {
		// run test
		test.InternalRunSubTests(ctx, t)
		// Note: in case of failed tests, we don't return error. GO's testing framework should be able to figure it out from t.Failed()
		return nil
	}
}

func testSetup(ctx context.Context, t *testing.T, hooks []test.Hook) context.Context {
	// run setup hooks
	for _, h := range hooks {
		var e error
		ctx, e = h.Setup(ctx, t)
		if e != nil {
			t.Fatalf("error when setup test: %v", e)
		}
	}
	return ctx
}

func testTeardown(ctx context.Context, t *testing.T, hooks []test.Hook) {
	// register cleanup
	for i := len(hooks) - 1; i >= 0; i-- {
		if e := hooks[i].Teardown(ctx, t); e != nil {
			t.Fatalf("error when setup test: %v", e)
		}
	}
}

func mergeInitContext(sources ...context.Context) bootstrap.ContextOption {
	return func(ctx context.Context) context.Context {
		srcs := make([]context.Context, len(sources) + 1)
		srcs[0] = ctx
		for i := range sources {
			srcs[i+1] = sources[i]
		}
		return newMergedContext(srcs...)
	}
}

/************************
	Init Context
 ************************/

func newMergedContext(ctxList ...context.Context) context.Context {
	done := make(chan struct{})
	cases := make([]reflect.SelectCase, len(ctxList))
	for i, ctx := range ctxList {
		cases[i] = reflect.SelectCase{
			Dir:  reflect.SelectRecv,
			Chan: reflect.ValueOf(ctx.Done()),
		}
	}
	go func() {
		_, _, _ = reflect.Select(cases)
		close(done)
	}()

	return &mergedContext{
		sources: ctxList,
		done:    done,
	}
}

type mergedContext struct {
	sources []context.Context
	done    <-chan struct{}
}

func (mc mergedContext) Deadline() (earliest time.Time, ok bool) {
	for _, ctx := range mc.sources {
		if deadline, subOk := ctx.Deadline(); subOk && (earliest.IsZero() || deadline.Before(earliest)) {
			earliest = deadline
			ok = true
		}
	}
	return
}

func (mc mergedContext) Done() <-chan struct{} {
	return mc.done
}

func (mc mergedContext) Err() error {
	for _, ctx := range mc.sources {
		if err := ctx.Err(); err != nil {
			return err
		}
	}
	return nil
}

func (mc mergedContext) Value(key any) any {
	for _, ctx := range mc.sources {
		if v := ctx.Value(key); v != nil {
			return v
		}
	}
	return nil
}
