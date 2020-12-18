package requestcache

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/session"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"encoding/gob"
	"go.uber.org/fx"
)

var Module = &bootstrap.Module{
	Name: "session",
	Precedence: security.MinSecurityPrecedence + 40, //not really important, because no invoke is added here
	Options: []fx.Option{
		fx.Provide(newRequestCacheAccessor),
	},
}

func init() {
	bootstrap.Register(Module)
	gob.Register((*Request)(nil))
}

func newRequestCacheAccessor(sessionStore session.Store) web.RequestCacheAccessor {
	m := &Accessor{
		store: sessionStore,
	}
	return m
}