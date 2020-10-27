package store

import (
	"cto-github.cisco.com/livdu/jupiter/pkg/bootstrap"
	"cto-github.cisco.com/livdu/jupiter/pkg/security/session"
	"errors"
	"go.uber.org/fx"
)

var Module = &bootstrap.Module{
	Precedence: -1,
	Options: []fx.Option{
		fx.Provide(NewSessionStore),
	},
}

func init() {
	bootstrap.Register(Module)
}

// Maker func, does nothing. Allow service to include this module in main()
func Use() {

}

/**************************
	Provider
***************************/
func NewSessionStore(ctx *bootstrap.ApplicationContext) (session.Store, error) {
	var secret []byte
	switch v := ctx.Value("security.session.secret"); v.(type) {
	case string:
		secret = []byte(v.(string))
	default:
		return nil, errors.New("session secret not available, set security.session.secret please")
	}

	// TODO create different type based on properties
	return NewMemoryStore(secret), nil
}

