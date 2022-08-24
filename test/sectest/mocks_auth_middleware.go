package sectest

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"github.com/gin-gonic/gin"
)

type MockAuthenticationMiddleware struct {
	MWMocker             MWMocker
}

func NewMockAuthenticationMiddleware(authentication security.Authentication) *MockAuthenticationMiddleware {
	return &MockAuthenticationMiddleware{
		MWMocker: MWMockFunc(func(MWMockContext) security.Authentication {
			return authentication
		}),
	}
}

func (m *MockAuthenticationMiddleware) AuthenticationHandlerFunc() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		auth := m.MWMocker.Mock(MWMockContext{
			Request: ctx.Request,
		})
		if auth != nil {
			ctx.Set(security.ContextKeySecurity, auth)
		}
	}
}

type MockUserAuthOptions func(opt *MockUserAuthOption)

type MockUserAuthOption struct {
	Principal   string
	Permissions map[string]interface{}
	State       security.AuthenticationState
	Details     interface{}
}

type mockUserAuthentication struct {
	Subject       string
	PermissionMap map[string]interface{}
	StateValue    security.AuthenticationState
	details       interface{}
}

func NewMockedUserAuthentication(opts ...MockUserAuthOptions) *mockUserAuthentication {
	opt := MockUserAuthOption{}
	for _, f := range opts {
		f(&opt)
	}
	return &mockUserAuthentication{
		Subject:       opt.Principal,
		PermissionMap: opt.Permissions,
		StateValue:    opt.State,
		details:       opt.Details,
	}
}

func (a *mockUserAuthentication) Principal() interface{} {
	return a.Subject
}

func (a *mockUserAuthentication) Permissions() security.Permissions {
	return a.PermissionMap
}

func (a *mockUserAuthentication) State() security.AuthenticationState {
	return a.StateValue
}

func (a *mockUserAuthentication) Details() interface{} {
	return a.details
}
