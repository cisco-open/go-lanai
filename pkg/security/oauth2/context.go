package oauth2

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
)

/******************************
	security.Authentication
******************************/
// Authentication implements security.Authentication
type Authentication interface {
	security.Authentication
	UserAuthentication() security.Authentication
	OAuth2Request() OAuth2Request
	AccessToken() AccessToken
}

type AuthenticationOptions func(opt *AuthOption)

type AuthOption struct {
	Request  OAuth2Request
	UserAuth security.Authentication
	Token    AccessToken
	Details  interface{}
}

// authentication
type authentication struct {
	Request   OAuth2Request                `json:"request"`
	UserAuth  security.Authentication      `json:"userAuth"`
	AuthState security.AuthenticationState `json:"state"`
	token     AccessToken
	details   interface{}
}

func NewAuthentication(opts...AuthenticationOptions) Authentication {
	config := AuthOption{}
	for _, opt := range opts {
		opt(&config)
	}
	return &authentication{
		Request:   config.Request,
		UserAuth:  config.UserAuth,
		AuthState: calculateState(config.Request, config.UserAuth),
		token:     config.Token,
		details:   config.Details,
	}
}

func (a *authentication) Principal() interface{} {
	if a.UserAuth == nil {
		return a.Request.ClientId()
	}
	return a.UserAuth.Principal()
}

func (a *authentication) Permissions() security.Permissions {
	if a.UserAuth == nil {
		return map[string]interface{}{}
	}
	return a.UserAuth.Permissions()
}

func (a *authentication) State() security.AuthenticationState {
	return a.AuthState
}

func (a *authentication) Details() interface{} {
	return a.details
}

func (a *authentication) UserAuthentication() security.Authentication {
	return a.UserAuth
}

func (a *authentication) OAuth2Request() OAuth2Request {
	return a.Request
}

func (a *authentication) AccessToken() AccessToken {
	return a.token
}

func calculateState(req OAuth2Request, userAuth security.Authentication) security.AuthenticationState {
	if req.Approved() {
		if userAuth != nil {
			return userAuth.State()
		}
		return security.StateAuthenticated
	} else if userAuth != nil {
		return security.StatePrincipalKnown
	}
	return security.StateAnonymous
}

/******************************
	UserAuthentication
******************************/
type UserAuthOptions func(opt *UserAuthOption)

type UserAuthOption struct {
	Principal   string
	Permissions map[string]interface{}
	State       security.AuthenticationState
	Details     map[string]interface{}
}

// userAuthentication implments security.Authentication.
// it represents basic information that could be typically extracted from JWT claims
type userAuthentication struct {
	Subject       string                       `json:"principal"`
	PermissionMap map[string]interface{}       `json:"permissions"`
	StateValue    security.AuthenticationState `json:"state"`
	DetailsMap    map[string]interface{}       `json:"details"`
}

func NewUserAuthentication(opts...UserAuthOptions) *userAuthentication {
	opt := UserAuthOption{}
	for _, f := range opts {
		f(&opt)
	}
	return &userAuthentication{
		Subject:       opt.Principal,
		PermissionMap: opt.Permissions,
		StateValue:    opt.State,
		DetailsMap:    opt.Details,
	}
}

func (a *userAuthentication) Principal() interface{} {
	return a.Subject
}

func (a *userAuthentication) Permissions() security.Permissions {
	return a.PermissionMap
}

func (a *userAuthentication) State() security.AuthenticationState {
	return a.StateValue
}

func (a *userAuthentication) Details() interface{} {
	return a.DetailsMap
}



