package sectest

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"github.com/gin-gonic/gin"
)



type MockAuthenticationMiddleware struct {
	MockedAuthentication security.Authentication
}

func NewMockAuthenticationMiddleware(authentication security.Authentication) *MockAuthenticationMiddleware {
	return &MockAuthenticationMiddleware{
		MockedAuthentication: authentication,
	}
}

func (m *MockAuthenticationMiddleware) AuthenticationHandlerFunc() gin.HandlerFunc{
	return func(ctx *gin.Context) {
		ctx.Set(security.ContextKeySecurity, m.MockedAuthentication)
	}
}

type MockUserAuthOptions func(opt *MockUserAuthOption)

type MockUserAuthOption struct {
	Principal   interface{}
	Permissions map[string]interface{}
	State       security.AuthenticationState
	Details     interface{}
}

type mockUserAuthentication struct {
	Subject       interface{}
	PermissionMap map[string]interface{}
	StateValue    security.AuthenticationState
	details       interface{}
}

func NewMockedUserAuthentication(opts...MockUserAuthOptions) *mockUserAuthentication {
	opt := MockUserAuthOption{}
	for _, f := range opts {
		f(&opt)
	}
	return &mockUserAuthentication{
		Subject:       opt.Principal,
		PermissionMap: opt.Permissions,
		StateValue:    opt.State,
		details:    opt.Details,
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