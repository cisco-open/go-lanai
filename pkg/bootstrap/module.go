package bootstrap

import (
	"go.uber.org/fx"
)

const (
	LowestPrecedence  = int(^uint(0) >> 1)    // max int
	HighestPrecedence = -LowestPrecedence - 1 // min int

	FrameworkModulePrecedenceBandwidth = 799
	FrameworkModulePrecedence          = LowestPrecedence - 200 * (FrameworkModulePrecedenceBandwidth+ 1)
	AnonymousModulePrecedence          = FrameworkModulePrecedence - 1
	PriorityModulePrecedence           = HighestPrecedence + 1
)

const (
	_ = FrameworkModulePrecedence + iota * (FrameworkModulePrecedenceBandwidth+ 1)
	AppConfigPrecedence
	TracingPrecedence
	ActuatorPrecedence
	ConsulPrecedence
	VaultPrecedence
	RedisPrecedence
	DatabasePrecedence
	KafkaPrecedence
	WebPrecedence
	SecurityPrecedence
	DebugPrecedence
	ServiceDiscoveryPrecedence
	DistributedLockPrecedence
	TenantHierarchyAccessorPrecedence
	TenantHierarchyLoaderPrecedence
	TenantHierarchyModifierPrecedence
	HttpClientPrecedence
	SecurityIntegrationPrecedence
	SwaggerPrecedence
	StartupSummaryPrecedence
	CommandLineRunnerPrecedence
	MigrationPrecedence //migration's invocation should happen after everything else, in case it needs functionality from any other modules
)

type Module struct {
	// Precedence basically govern the order or invokers between different Bootstrapper
	Name            string
	Precedence      int
	PriorityOptions []fx.Option
	Options         []fx.Option
}

func newAnonymousModule() *Module {
	return &Module{
		Name:       "anonymous",
		Precedence: AnonymousModulePrecedence,
	}
}

func newApplicationMainModule() *Module {
	return &Module{
		Name:       "main",
		Precedence: PriorityModulePrecedence,
	}
}
