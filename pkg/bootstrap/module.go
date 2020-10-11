package bootstrap

import (
	"go.uber.org/fx"
	"sync"
)

const (
	LowestPrecedence = int(^uint(0) >> 1) // max int
	HighestPrecedence = -LowestPrecedence - 1 // min int
)

var anonymousOnce sync.Once
var anonymous *Module

type Module struct {
	// Precedence basically govern the order or invokers between different Bootstrapper
	Precedence int
	Provides []fx.Option
	Invokes  []fx.Option
}

func anonymousModule() *Module {
	anonymousOnce.Do(func() {
		anonymous = &Module{
			Precedence: LowestPrecedence,
		}
	})
	return anonymous
}