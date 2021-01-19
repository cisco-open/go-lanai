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

type AuthenticationOptionsFunc func(*AuthConfig)

type AuthConfig struct {
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

func NewAuthentication(opts...AuthenticationOptionsFunc) Authentication {
	config := AuthConfig{}
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

func (a *authentication) Permissions() map[string]interface{} {
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

