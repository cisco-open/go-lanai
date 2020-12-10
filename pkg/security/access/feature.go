package access

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"fmt"
)

//goland:noinspection GoNameStartsWithPackageName
type AccessControl struct {
	owner *AccessControlFeature
	order int
	matcher AcrMatcher
	control ControlFunc
}

// order.Ordered
func (ac *AccessControl) Order() int {
	return ac.order
}

func (ac *AccessControl) WithOrder(order int) *AccessControl {
	ac.order = order
	return ac
}

func (ac *AccessControl) PermitAll() *AccessControlFeature {
	ac.control = PermitAll
	return ac.owner
}

func (ac *AccessControl) DenyAll() *AccessControlFeature {
	ac.control = DenyAll
	return ac.owner
}

func (ac *AccessControl) Authenticated() *AccessControlFeature {
	ac.control = Authenticated
	return ac.owner
}

func (ac *AccessControl) HasPermissions(permissions...string) *AccessControlFeature {
	ac.control = HasPermissions(permissions...)
	return ac.owner
}

func (ac *AccessControl) AllowIf(cf ControlFunc) *AccessControlFeature {
	ac.control = cf
	return ac.owner
}

//goland:noinspection GoNameStartsWithPackageName
type AccessControlFeature struct {
	acl []*AccessControl
}

// Standard security.Feature entrypoint
func (f *AccessControlFeature) Identifier() security.FeatureIdentifier {
	return FeatureId
}

// AccessControlFeature specifics
func (f *AccessControlFeature) Request(matcher AcrMatcher) *AccessControl {
	ac := &AccessControl{
		owner: f,
		matcher: matcher,
	}
	f.acl = append(f.acl, ac)
	return ac
}

func Configure(ws security.WebSecurity) *AccessControlFeature {
	feature := New()
	if fc, ok := ws.(security.FeatureModifier); ok {
		return  fc.Enable(feature).(*AccessControlFeature)
	}
	panic(fmt.Errorf("unable to configure access control: provided WebSecurity [%T] doesn't support FeatureModifier", ws))
}

// Standard security.Feature entrypoint, DSL style. Used with security.WebSecurity
func New() *AccessControlFeature {
	return &AccessControlFeature{}
}

/**************************
	Common ControlFunc
***************************/
func PermitAll(_ security.Authentication) (bool, error) {
	return true, nil
}

func DenyAll(_ security.Authentication) (bool, error) {
	return false, nil
}

func Authenticated(auth security.Authentication) (bool, error) {
	if auth.State() >= security.StateAuthenticated {
		return true, nil
	}
	return false, security.NewInsufficientAuthError("not fully authenticated")
}

func HasPermissions(permissions...string) ControlFunc {
	return func(auth security.Authentication) (bool, error) {
		switch {
		case auth.State() > security.StateAnonymous && security.HasPermissions(auth, permissions...):
			return true, nil
		case auth.State() < security.StatePrincipalKnown:
			return false, security.NewInsufficientAuthError("not authenticated")
		case auth.State() < security.StateAuthenticated:
			return false, security.NewInsufficientAuthError("not fully authenticated")
		default:
			return false, security.NewAccessDeniedError("access denied")
		}
	}
}
