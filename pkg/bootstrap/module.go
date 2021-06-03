package bootstrap

import (
	"go.uber.org/fx"
	"sync"
)

const (
	LowestPrecedence  = int(^uint(0) >> 1)    // max int
	HighestPrecedence = -LowestPrecedence - 1 // min int

	FrameworkModulePrecedenceBandwith = 799
	FrameworkModulePrecedence         = LowestPrecedence - 200 * (FrameworkModulePrecedenceBandwith + 1)
	AnonymousModulePrecedence         = FrameworkModulePrecedence - 1
	PriorityModulePrecedence          = HighestPrecedence + 1
)

const (
	_ = FrameworkModulePrecedence + iota * (FrameworkModulePrecedenceBandwith + 1)
	AppConfigPrecedence
	TracingPrecedence
	ActuatorPrecedence
	ConsulPrecedence
	VaultPrecedence
	RedisPrecedence
	DatabasePrecedence
	WebPrecedence
	SecurityPrecedence
	ServiceDiscoveryPrecedence
	TenantHierarchyAccessorPrecedence
	TenantHierarchyLoaderPrecedence
	TenantHierarchyModifierPrecedence
	HttpClientPrecedence
	SecurityIntegrationPrecedence
	SwaggerPrecedence
	CommandLineRunnerPrecedence
	MigrationPrecedence //migration's invocation should happen after everything else, in case it needs functionality from any other modules
)

var anonymousOnce sync.Once
var anonymous *Module

var applicationMainOnce sync.Once
var applicationMain *Module

type Module struct {
	// Precedence basically govern the order or invokers between different Bootstrapper
	Name            string
	Precedence      int
	PriorityOptions []fx.Option
	Options         []fx.Option
}

func anonymousModule() *Module {
	anonymousOnce.Do(func() {
		anonymous = &Module{
			Name:       "anonymous",
			Precedence: AnonymousModulePrecedence,
		}
	})
	return anonymous
}

func applicationMainModule() *Module {
	applicationMainOnce.Do(func() {
		applicationMain = &Module{
			Name:       "main",
			Precedence: PriorityModulePrecedence,
		}
	})
	return applicationMain
}
