package security

import (
	"context"
	"encoding/gob"
	"errors"
	"github.com/gin-gonic/gin"
)

const (
	HighestMiddlewareOrder = int(- 1 << 18 + 1) // -0x3ffff = -262143
	LowestMiddlewareOrder = HighestMiddlewareOrder + 0xffff // -0x30000 = -196608
)

type AuthenticationState int
const (
	StateAnonymous = AuthenticationState(iota)
	StatePrincipalKnown
	StateAuthenticated
)

type Authentication interface {
	Principal() interface{}
	Permissions() map[string]interface{}
	State() AuthenticationState
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

func (EmptyAuthentication) State() AuthenticationState {
	return StateAnonymous
}

func (EmptyAuthentication) Permissions() map[string]interface{} {
	return map[string]interface{}{}
}

func GobRegister() {
	gob.Register(EmptyAuthentication(""))
	gob.Register((*AnonymousAuthentication)(nil))
	gob.Register((*codedError)(nil))
	gob.Register((*nestedError)(nil))
	gob.Register(errors.New(""))
}

func Get(ctx context.Context) Authentication {
	secCtx, ok := ctx.Value(ContextKeySecurity).(Authentication)
	if !ok {
		secCtx = EmptyAuthentication("EmptyAuthentication")
	}
	return secCtx
}

func Clear(ctx *gin.Context) {
	ctx.Set(gin.AuthUserKey, nil)
	ctx.Set(ContextKeySecurity, nil)
}

// TryClear attempt to clear security context. Return true if succeeded
func TryClear(ctx context.Context) bool {
	switch ctx.(type) {
	case *gin.Context:
		Clear(ctx.(*gin.Context))
	default:
		return false
	}
	return true
}

func HasPermissions(auth Authentication, permissions...string) bool {
	for _,p := range permissions {
		_, ok := auth.Permissions()[p]
		if !ok {
			return false
		}
	}
	return true
}

func IsFullyAuthenticated(auth Authentication) bool {
	return auth.State() >= StateAuthenticated
}

func IsBeingAuthenticated(from, to Authentication) bool {
	fromUnauthenticatedState := from == nil || from.State() < StateAuthenticated
	toAuthenticatedState := to != nil && to.State() > StatePrincipalKnown
	return fromUnauthenticatedState && toAuthenticatedState
}

func IsBeingUnAuthenticated(from, to Authentication) bool {
	fromAuthenticated := from != nil && from.State() > StateAnonymous
	toUnAuthenticatedState := to == nil || to.State() <= StateAnonymous
	return fromAuthenticated && toUnAuthenticatedState
}