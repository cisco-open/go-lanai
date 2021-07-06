package suitetest

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/order"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"fmt"
	"os"
	"sort"
	"testing"
)

const (
	HookOrderPackage int = - 0xffff
)

type PackageHook interface {
	Setup() error
	Teardown() error
}

type PackageOptions func(opt *pkg)

type pkg struct {
	PackageHooks []PackageHook
	TestOptions  []test.Options
}

func RunTests(m *testing.M, opts ...PackageOptions) {
	s := pkg{
		PackageHooks: []PackageHook{},
		TestOptions:  []test.Options{},
	}
	for _, fn := range opts {
		fn(&s)
	}
	sort.SliceStable(s.PackageHooks, func(i, j int) bool {
		return order.OrderedFirstCompare(s.PackageHooks[i], s.PackageHooks[j])
	})

	// run setup TestHooks
	for _, h := range s.PackageHooks {
		if e := h.Setup(); e != nil {
			panic(fmt.Errorf("error when setup test pkg: %v", e))
		}
	}

	// register DefaultTestHook
	test.InternalOptions = s.TestOptions

	// run tests
	code := m.Run()
	// run teardown TestHooks in reversed order
	for i := len(s.PackageHooks) - 1; i >= 0; i-- {
		if e := s.PackageHooks[i].Teardown(); e != nil {
			panic(fmt.Errorf("error when teardown test pkg: %v", e))
		}
	}

	os.Exit(code)
}

/****************************
	Suite Options
 ****************************/

// SetupFunc is package level setup function that run once per package
type SetupFunc func() error
// TeardownFunc is package level teardown function that run once per package
type TeardownFunc func() error

// orderedSuiteHook implements PackageHook and order.Ordered
type orderedSuiteHook struct {
	order        int
	setupFunc    SetupFunc
	teardownFunc TeardownFunc
}

func (h *orderedSuiteHook) Order() int {
	return h.order
}

func (h *orderedSuiteHook) Setup() error {
	if h.setupFunc == nil {
		return nil
	}
	return h.setupFunc()
}

func (h *orderedSuiteHook) Teardown() error {
	if h.teardownFunc == nil {
		return nil
	}
	return h.teardownFunc()
}

// WithOptions group multiple PackageOptions into one, typically used for other test utilities to provide
// single entry point of certain feature.
// Not recommended for test implementers to use directly
func WithOptions(opts ...PackageOptions) PackageOptions {
	return func(opt *pkg) {
		for _, fn := range opts {
			fn(opt)
		}
	}
}

// Setup register the given setup function to run at order 0, higher(smaller) order runs first
// package setup runs once per test package, and should only be registered in TestMain(m *testing.M)
func Setup(fn SetupFunc) PackageOptions {
	return SetupWithOrder(0, fn)
}

// SetupWithOrder register the given setup function to run at given order, higher(smaller) order runs first
// package setup runs once per test package, and should only be registered in TestMain(m *testing.M)
func SetupWithOrder(order int, fn SetupFunc) PackageOptions {
	return func(opt *pkg) {
		opt.PackageHooks = append(opt.PackageHooks, &orderedSuiteHook{
			order:     order,
			setupFunc: fn,
		})
	}
}

// Teardown register the given teardown function to run at order 0, higher(smaller) order runs LAST
// package teardown runs once per test package, and should only be registered in TestMain(m *testing.M)
func Teardown(fn TeardownFunc) PackageOptions {
	return TeardownWithOrder(0, fn)
}

// TeardownWithOrder register the given teardown function to run at given order, higher(smaller) order runs LAST
// package teardown runs once per test package, and should only be registered in TestMain(m *testing.M)
func TeardownWithOrder(order int, fn TeardownFunc) PackageOptions {
	return func(opt *pkg) {
		opt.PackageHooks = append(opt.PackageHooks, &orderedSuiteHook{
			order:        order,
			teardownFunc: fn,
		})
	}
}

// TestOptions register per-test options at package level: only declared once in TestMain(m *testing.M)
// All test.Options are applied once per Test*()
func TestOptions(opts ...test.Options) PackageOptions {
	return func(opt *pkg) {
		opt.TestOptions = append(opt.TestOptions, opts...)
	}
}

// TestSetup is a convenient function equivalent to TestOptions(test.Setup(fn))
func TestSetup(fn test.SetupFunc) PackageOptions {
	return TestSetupWithOrder(HookOrderPackage, fn)
}

// TestSetupWithOrder is a convenient function equivalent to TestOptions(test.Hooks(test.NewSetupHook(order, fn)))
func TestSetupWithOrder(order int, fn test.SetupFunc) PackageOptions {
	return TestOptions(test.Hooks(test.NewSetupHook(order, fn)))
}

// TestTeardown is a convenient function equivalent to TestOptions(test.Teardown(fn))
func TestTeardown(fn test.TeardownFunc) PackageOptions {
	return TestTeardownWithOrder(HookOrderPackage, fn)
}

// TestTeardownWithOrder is a convenient function equivalent to TestOptions(test.Hooks(test.NewTeardownHook(order, fn)))
func TestTeardownWithOrder(order int, fn test.TeardownFunc) PackageOptions {
	return TestOptions(test.Hooks(test.NewTeardownHook(order, fn)))
}
