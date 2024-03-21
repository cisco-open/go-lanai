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

package bootstrap

import (
	"go.uber.org/fx"
)

const (
	LowestPrecedence  = int(^uint(0) >> 1)    // max int
	HighestPrecedence = -LowestPrecedence - 1 // min int

	FrameworkModulePrecedenceBandwidth = 799
	FrameworkModulePrecedence          = LowestPrecedence - 200*(FrameworkModulePrecedenceBandwidth+1)
	AnonymousModulePrecedence          = FrameworkModulePrecedence - 1
	PriorityModulePrecedence           = HighestPrecedence + 1
)

const (
	_ = FrameworkModulePrecedence + iota*(FrameworkModulePrecedenceBandwidth+1)
	AppConfigPrecedence
	TracingPrecedence
	ActuatorPrecedence
	ConsulPrecedence
	VaultPrecedence
	AwsPrecedence
	TlsConfigPrecedence
	RedisPrecedence
	DatabasePrecedence
	KafkaPrecedence
	OpenSearchPrecedence
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
	// CommandLineRunnerPrecedence invocation should happen after everything else, in case it needs functionality from any other modules
	CommandLineRunnerPrecedence
)

type Module struct {
	Name            string
	// Precedence basically govern the order or invokers between different Modules
	Precedence      int
	// PriorityOptions are fx.Options applied before any regular Options
	PriorityOptions []fx.Option
	// Options is a collection fx.Option: fx.Provide, fx.Invoke, etc.
	Options         []fx.Option
	// Modules is a collection of *Module that will also be initialized.
	// They are not necessarily sub-modules. During bootstrapping, all modules are flattened and Precedence are calculated at the end
	Modules 		[]*Module
}

// newAnonymousModule has lower precedence than framework modules.
// It is used to hold options registered via AddOptions()
func newAnonymousModule() *Module {
	return &Module{
		Name:       "anonymous",
		Precedence: AnonymousModulePrecedence,
	}
}

// newApplicationMainModule application main module has the highest precedence.
// It is used to hold options passed in via NewAppCmd()
func newApplicationMainModule() *Module {
	return &Module{
		Name:       "main",
		Precedence: PriorityModulePrecedence,
	}
}
