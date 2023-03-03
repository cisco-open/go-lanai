// Package v1 Generated by lanai_cli codegen. DO NOT EDIT
package v1

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"go.uber.org/fx"
)

var Module = &bootstrap.Module{
	Name:       "v1-controller",
	Precedence: bootstrap.AnonymousModulePrecedence,
	Options: []fx.Option{
		web.FxControllerProviders(
			NewRequestBodyTestsIdController,
			NewTestpathScopeController,
			NewUuidtestIdController,
		),
	},
}
