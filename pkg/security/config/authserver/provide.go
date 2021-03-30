package authserver

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/redis"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/auth"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/auth/revoke"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/common"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/session"
	"go.uber.org/fx"
)

type initDI struct {
	fx.In
	AppContext         *bootstrap.ApplicationContext
	RedisClientFactory redis.ClientFactory
	SessionStore       session.Store
}

type initOut struct {
	fx.Out
	ContextDetailsStore security.ContextDetailsStore
	AuthRegistry        auth.AuthorizationRegistry
	AccessRevoker       auth.AccessRevoker
}

func provide(di initDI) initOut {
	store := common.NewRedisContextDetailsStore(di.AppContext, di.RedisClientFactory)
	revoker := revoke.NewDefaultAccessRevoker(func(opt *revoke.RevokerOption) {
		opt.AuthRegistry = store
		opt.SessionStore = di.SessionStore
	})
	return initOut{
		ContextDetailsStore: store,
		AuthRegistry: store,
		AccessRevoker: revoker,
	}
}
