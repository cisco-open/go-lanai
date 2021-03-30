package authserver

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/auth"
	"go.uber.org/fx"
)

type provideDI struct {
	fx.In
	Config         *Configuration
}

type provideOut struct {
	fx.Out
	AccessRevoker       auth.AccessRevoker
}

func provide(di provideDI) provideOut {
	return provideOut{
		AccessRevoker: di.Config.accessRevoker(),
	}
}
