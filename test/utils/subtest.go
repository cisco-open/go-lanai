package test

import (
	"context"
	"fmt"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"math/rand"
	"path"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

/****************************
	Sub Tests
 ****************************/

type SubTestFunc func(ctx context.Context, t *testing.T)
type GomegaSubTestFunc func(ctx context.Context, t *testing.T, g *gomega.WithT)

func SubTestFuncWithGomega(st GomegaSubTestFunc) SubTestFunc {
	return func(ctx context.Context, t *testing.T) {
		st(ctx, t, NewWithT(t))
	}
}

// FuncName returns a name that could potentially used as sub test name
// function panic if given fn is not func
func FuncName(fn interface{}, suffixed bool) string {
	fnName := runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).Name()
	_, fnName = path.Split(fnName)
	// we assume fnName is in format of "<package>[.receiver]<.NamedFunction>[.funcN[<.N>...]]
	// what we want is "NamedFunction"
	fnName = strings.SplitN(fnName, ".func", 2)[0]
	split := strings.Split(fnName, ".")
	fnName = split[len(split) - 1]

	// remove "-fm"
	fnName = strings.TrimSuffix(fnName, "-fm")

	if suffixed {
		fnName = fnName + "@" + randString(4)
	}
	return fnName
}


/****************************
	Test Options
 ****************************/

func SubTest(name string, subtest SubTestFunc) Options {
	return func(opt *T) {
		opt.SubTests[name] = subtest
	}
}

func AnonymousSubTest(st SubTestFunc) Options {
	return SubTest(FuncName(st, true), st)
}

func GomegaSubTest(st GomegaSubTestFunc, name ...string) Options {
	var n string
	if len(name) > 0 {
		args := make([]interface{}, len(name)-1)
		for i, v := range name[1:] {
			args[i] = v
		}
		n = fmt.Sprintf(name[0], args...)
	} else {
		n = FuncName(st, true)
	}
	return SubTest(n, SubTestFuncWithGomega(st))
}

func SubTestSetup(fn SetupFunc) Options {
	return func(opt *T) {
		opt.SubTestHooks = append(opt.SubTestHooks, &orderedHook{
			setupFunc: fn,
		})
	}
}

func SubTestTeardown(fn TeardownFunc) Options {
	return func(opt *T) {
		opt.SubTestHooks = append(opt.SubTestHooks, &orderedHook{
			teardownFunc: fn,
		})
	}
}

/****************************
	Test Options
 ****************************/

func randString(length int) string {
	const charset ="0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}