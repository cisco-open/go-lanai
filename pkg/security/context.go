package security

import "context"

const (
	ContextKeySecurity = "kSecurity"
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

func (EmptyAuthentication) MoreActionRequired() bool {
	return false
}

func Get(ctx context.Context) Authentication {
	secCtx, ok := ctx.Value(ContextKeySecurity).(Authentication)
	if !ok {
		secCtx = EmptyAuthentication("EmptyAuthentication")
	}
	return secCtx
}
