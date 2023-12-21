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

package test

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/order"
	"sort"
	"testing"
)

var (
	// InternalOptions is internal variable, exported for cross-package access
	// InternalOptions holds common setup/teardown hooks of all tests in same package.
	// testsuite package has options to set this list
	// Note, when executing all tests, golang run tests on per-package basis
	InternalOptions = make([]Options, 0)
)

// InternalRunner is an internal type, exported for cross-package reference
type InternalRunner func(context.Context, *T)

type SetupFunc func(ctx context.Context, t *testing.T) (context.Context, error)
type TeardownFunc func(ctx context.Context, t *testing.T) error

// Hook is registered for tests and sub tests, should provide SetupFunc or TeardownFunc (or both)
// This interface is mostly internal usage.
// Test implementers typically use Options to create instance of this interface
type Hook interface {
	Setup(ctx context.Context, t *testing.T) (context.Context, error)
	Teardown(ctx context.Context, t *testing.T) error
}

// Options are test config functions to pass into RunTest
type Options func(opt *T)

// T embed *testing.T and holds additional information of test config
type T struct {
	*testing.T
	runner       InternalRunner
	TestHooks    []Hook
	SubTestHooks []Hook
	SubTests     *SubTestOrderedMap
}

// RunTest is the entry point of any Test...().
// It takes any context, and run sub tests according to provided Options
func RunTest(ctx context.Context, t *testing.T, opts ...Options) {
	test := T{
		T:            t,
		runner:       unitTestRunner,
		TestHooks:    []Hook{},
		SubTestHooks: []Hook{},
		SubTests:     NewSubTestOrderedMap(),
	}
	for _, fn := range InternalOptions {
		fn(&test)
	}
	for _, fn := range opts {
		fn(&test)
	}

	sort.SliceStable(test.TestHooks, func(i, j int) bool {
		return order.OrderedFirstCompare(test.TestHooks[i], test.TestHooks[j])
	})

	test.runner(ctx, &test)
}

func unitTestRunner(ctx context.Context, t *T) {
	// run setup TestHooks
	ctx = runTestSetupHooks(ctx, t.T, t.TestHooks, "error when setup test")
	defer runTestTeardownHooks(ctx, t.T, t.TestHooks, "error when cleanup test")

	// run test
	InternalRunSubTests(ctx, t)
}

// InternalRunSubTests is an internal function. exported for cross-package reference
func InternalRunSubTests(ctx context.Context, t *T) {
	names := t.SubTests.Keys()
	for _, n := range names {
		if fn, ok := t.SubTests.Get(n); ok {
			t.Run(n, func(goT *testing.T) {
				ctx = runTestSetupHooks(ctx, goT, t.SubTestHooks, "error when setup sub test")
				defer runTestTeardownHooks(ctx, goT, t.SubTestHooks, "error when cleanup sub test")
				fn(ctx, goT)
			})
		}
	}
}

func runTestSetupHooks(ctx context.Context, t *testing.T, hooks []Hook, errMsg string) context.Context {
	// run setup TestHooks
	for _, h := range hooks {
		var e error
		ctx, e = h.Setup(ctx, t)
		if e != nil {
			t.Fatalf("%s: %v", errMsg, e)
		}
	}
	return ctx
}

func runTestTeardownHooks(ctx context.Context, t *testing.T, hooks []Hook, errMsg string) {
	for _, h := range hooks {
		if e := h.Teardown(ctx, t); e != nil {
			t.Fatalf("%s: %v", errMsg, e)
		}
	}
}

//func setupCleanup(t *testing.T, hooks []Hook, errMsg string) {
//	// register cleanup
//	for _, h := range hooks {
//		fn := h.Teardown
//		t.Cleanup(func() {
//			if e := fn(t); e != nil {
//				t.Fatalf("%s: %v", errMsg, e)
//			}
//		})
//	}
//}

// orderedHook implements Hook and order.Ordered
type orderedHook struct {
	order        int
	setupFunc    SetupFunc
	teardownFunc TeardownFunc
}

func NewHook(order int, setupFunc SetupFunc, teardownFunc TeardownFunc) *orderedHook {
	return &orderedHook{
		order:        order,
		setupFunc:    setupFunc,
		teardownFunc: teardownFunc,
	}
}

func NewSetupHook(order int, setupFunc SetupFunc) *orderedHook {
	return NewHook(order, setupFunc, nil)
}

func NewTeardownHook(order int, teardownFunc TeardownFunc) *orderedHook {
	return NewHook(order, nil, teardownFunc)
}

func (h *orderedHook) Order() int {
	return h.order
}

func (h *orderedHook) Setup(ctx context.Context, t *testing.T) (context.Context, error) {
	if h.setupFunc == nil {
		return ctx, nil
	}
	return h.setupFunc(ctx, t)
}

func (h *orderedHook) Teardown(ctx context.Context, t *testing.T) error {
	if h.teardownFunc == nil {
		return nil
	}
	return h.teardownFunc(ctx, t)
}

/****************************
	Common Test Options
 ****************************/

// WithInternalRunner is internal option, exported for cross-platform access
func WithInternalRunner(runner InternalRunner) Options {
	return func(opt *T) {
		opt.runner = runner
	}
}

// WithOptions group multiple options into one.
// This is mostly used by other testing utilities to provide grouped test configs
func WithOptions(opts ...Options) Options {
	return func(opt *T) {
		for _, fn := range opts {
			fn(opt)
		}
	}
}

// Setup is an Options that register the SetupFunc to run before ANY sub tests starts
func Setup(fn SetupFunc) Options {
	return func(opt *T) {
		opt.TestHooks = append(opt.TestHooks, NewSetupHook(0, fn))
	}
}

// Teardown is an Options that register the TeardownFunc to run after ALL sub tests finishs
func Teardown(fn TeardownFunc) Options {
	return func(opt *T) {
		opt.TestHooks = append(opt.TestHooks, NewTeardownHook(0, fn))
	}
}

// Hooks is an Options that register multiple Hook.
// Test implementers are recommended to use Setup or Teardown instead
func Hooks(hooks ...Hook) Options {
	return func(opt *T) {
		opt.TestHooks = append(opt.TestHooks, hooks...)
	}
}
