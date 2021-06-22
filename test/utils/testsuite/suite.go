package testsuite

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/order"
	test "cto-github.cisco.com/NFV-BU/go-lanai/test/utils"
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

type setupFunc func() error
type teardownFunc func() error

// orderedSuiteHook implements PackageHook and order.Ordered
type orderedSuiteHook struct {
	order        int
	setupFunc    setupFunc
	teardownFunc teardownFunc
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

func Setup(fn setupFunc) PackageOptions {
	return SetupWithOrder(0, fn)
}

func SetupWithOrder(order int, fn setupFunc) PackageOptions {
	return func(opt *pkg) {
		opt.PackageHooks = append(opt.PackageHooks, &orderedSuiteHook{
			order:     order,
			setupFunc: fn,
		})
	}
}

func Teardown(fn teardownFunc) PackageOptions {
	return TeardownWithOrder(0, fn)
}

func TeardownWithOrder(order int, fn teardownFunc) PackageOptions {
	return func(opt *pkg) {
		opt.PackageHooks = append(opt.PackageHooks, &orderedSuiteHook{
			order:        order,
			teardownFunc: fn,
		})
	}
}

func TestOptions(opts ...test.Options) PackageOptions {
	return func(opt *pkg) {
		opt.TestOptions = append(opt.TestOptions, opts...)
	}
}

func TestSetup(fn test.SetupFunc) PackageOptions {
	return TestSetupWithOrder(HookOrderPackage, fn)
}

func TestSetupWithOrder(order int, fn test.SetupFunc) PackageOptions {
	return TestOptions(test.Hooks(test.NewSetupHook(order, fn)))
}

func TestTeardown(fn test.TeardownFunc) PackageOptions {
	return TestTeardownWithOrder(HookOrderPackage, fn)
}

func TestTeardownWithOrder(order int, fn test.TeardownFunc) PackageOptions {
	return TestOptions(test.Hooks(test.NewTeardownHook(order, fn)))
}
