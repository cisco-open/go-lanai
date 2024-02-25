// Package v1 Generated by lanai-cli codegen. DO NOT EDIT
package v1

import (
	"github.com/cisco-open/go-lanai/pkg/bootstrap"
	"github.com/cisco-open/go-lanai/pkg/web"
	"go.uber.org/fx"
)

var Module = &bootstrap.Module{
	Name:       "v1-controller",
	Precedence: bootstrap.AnonymousModulePrecedence,
	Options: []fx.Option{
		web.FxControllerProviders(
			NewExampleFriendsController,
		),
	},
}
