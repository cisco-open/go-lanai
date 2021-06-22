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
type TeardownFunc func(t *testing.T) error

type Hook interface {
	Setup(ctx context.Context, t *testing.T) (context.Context, error)
	Teardown(t *testing.T) error
}

type Options func(opt *T)
type T struct {
	*testing.T
	runner       InternalRunner
	TestHooks    []Hook
	SubTestHooks []Hook
	SubTests     map[string]SubTestFunc
}

func RunTest(ctx context.Context, t *testing.T, opts ...Options) {
	test := T{
		T:            t,
		runner:       unitTestRunner,
		TestHooks:    []Hook{},
		SubTestHooks: []Hook{},
		SubTests:     map[string]SubTestFunc{},
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
	defer runTestTeardownHooks(t.T, t.TestHooks, "error when cleanup test")

	// run test
	InternalRunSubTests(ctx, t)
}

// InternalRunSubTests is an internal function. exported for cross-package reference
func InternalRunSubTests(ctx context.Context, t *T) {
	for n, fn := range t.SubTests {
		t.Run(n, func(goT *testing.T) {
			ctx = runTestSetupHooks(ctx, goT, t.SubTestHooks, "error when setup sub test")
			defer runTestTeardownHooks(goT, t.SubTestHooks, "error when cleanup sub test")
			fn(ctx, goT)
		})
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

func runTestTeardownHooks(t *testing.T, hooks []Hook, errMsg string) {
	for _, h := range hooks {
		if e := h.Teardown(t); e != nil {
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

func (h *orderedHook) Teardown(t *testing.T) error {
	if h.teardownFunc == nil {
		return nil
	}
	return h.teardownFunc(t)
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

func Setup(fn SetupFunc) Options {
	return func(opt *T) {
		opt.TestHooks = append(opt.TestHooks, NewSetupHook(0, fn))
	}
}

func Teardown(fn TeardownFunc) Options {
	return func(opt *T) {
		opt.TestHooks = append(opt.TestHooks, NewTeardownHook(0, fn))
	}
}

func Hooks(hooks ...Hook) Options {
	return func(opt *T) {
		opt.TestHooks = append(opt.TestHooks, hooks...)
	}
}
