package sectest

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
)

/**************************
	Function
 **************************/

type SecurityContextOptions func(opt *SecurityContextOption)

type SecurityContextOption struct {
	// Authentication override any other fields
	Authentication security.Authentication
}

// ContextWithSecurity set given SecurityContextOptions in the given context, returning the new context
func ContextWithSecurity(ctx context.Context, opts ...SecurityContextOptions) context.Context {
	opt := SecurityContextOption{}
	for _, fn := range opts {
		fn(&opt)
	}
	return &testScopeContext{
		Context: ctx,
		auth:    opt.Authentication,
	}
}

// WithMockedSecurity used to mock an oauth2.Authentication in the given context, returning a new context
func WithMockedSecurity(ctx context.Context, opts ...SecurityMockOptions) context.Context {
	return ContextWithSecurity(ctx, MockedAuthentication(opts...))
}

/**************************
	Options
 **************************/

func Authentication(auth security.Authentication) SecurityContextOptions {
	return func(opt *SecurityContextOption) {
		opt.Authentication = auth
	}
}

func MockedAuthentication(opts ...SecurityMockOptions) SecurityContextOptions {
	return func(opt *SecurityContextOption) {
		details := NewMockedSecurityDetails(opts...).(*mockedSecurityDetails)
		user := oauth2.NewUserAuthentication(func(opt *oauth2.UserAuthOption) {
			opt.Principal = details.Username()
			opt.State = security.StateAuthenticated
			opt.Permissions = map[string]interface{}{}
			for perm := range details.Permissions() {
				opt.Permissions[perm] = true
			}
		})
		token := &MockedToken{
			MockedTokenInfo: MockedTokenInfo{
				UName:       details.Username(),
				UID:         details.UserId(),
				TID:         details.TenantId(),
				TExternalId: details.TenantExternalId(),
				OrigU:       details.OrigUsername,
			},
			ExpTime: details.Exp,
			IssTime: details.Iss,
		}

		auth := oauth2.NewAuthentication(func(opt *oauth2.AuthOption) {
			opt.Request = oauth2.NewOAuth2Request(func(opt *oauth2.RequestDetails) {
				opt.ClientId = "mock"
				opt.Approved = true
			})
			opt.Token = token
			opt.UserAuth = user
			opt.Details = details
		})
		opt.Authentication = auth
	}
}

/**************************
	Context
 **************************/

// testScopeContext override security of parent context
type testScopeContext struct {
	context.Context
	auth security.Authentication
}

func (c testScopeContext) Value(key interface{}) interface{} {
	if key == security.ContextKeySecurity {
		return c.auth
	}
	return c.Context.Value(key)
}
