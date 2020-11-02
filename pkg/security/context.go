package security

import (
	"context"
	"encoding/gob"
)

const (
	HighestMiddlewareOrder = int(- 1 << 18 + 1) // -0x3ffff = -262143
	LowestMiddlewareOrder = HighestMiddlewareOrder + 0xffff // -0x30000 = -196608
)

type Authentication interface {
	Principal() interface{}
	Permissions() []string
	Authenticated() bool
	Details() interface{}
}

// EmptyAuthentication represent unauthenticated user.
// Note: anonymous user is considered authenticated
type EmptyAuthentication string

func (EmptyAuthentication) Principal() interface{} {
	return nil
}

func (EmptyAuthentication) Details() interface{} {
	return nil
}

func (EmptyAuthentication) Authenticated() bool {
	return false
}

func (EmptyAuthentication) Permissions() []string {
	return []string{}
}

func GobRegister() {
	gob.Register(EmptyAuthentication(""))
}

func Get(ctx context.Context) Authentication {
	secCtx, ok := ctx.Value(ContextKeySecurity).(Authentication)
	if !ok {
		secCtx = EmptyAuthentication("EmptyAuthentication")
	}
	return secCtx
}
