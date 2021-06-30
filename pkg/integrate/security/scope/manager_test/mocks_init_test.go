package manager_test

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/integrate/security/scope"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/integrate/security/seclient"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/sectest"
	"go.uber.org/fx"
)

/*************************
	Mocks
 *************************/

type mocksDIOut struct {
	fx.Out
	AuthClient   seclient.AuthenticationClient
	TokenReader  oauth2.TokenStoreReader
	TokenRevoker sectest.MockedTokenRevoker
	Counter InvocationCounter
}

func provideScopeMocksWithCounter(ctx *bootstrap.ApplicationContext) mocksDIOut {
	out := sectest.ProvideScopeMocks(ctx)
	counter := counter{
		AuthenticationClient: out.AuthClient,
		TokenStoreReader:     out.TokenReader,
		counts:               map[interface{}]*uint64{},
	}
	return mocksDIOut{
		AuthClient:   &counter,
		TokenReader:  &counter,
		TokenRevoker: out.TokenRevoker,
		Counter:      &counter,
	}
}

type noopScopeManager struct {}

func (m *noopScopeManager) StartScope(ctx context.Context, _ *scope.Scope) (context.Context, error) {
	return ctx, nil
}

func (m *noopScopeManager) Start(ctx context.Context, _ ...scope.Options) (context.Context, error) {
	return ctx, nil
}

func (m *noopScopeManager) End(ctx context.Context) context.Context {
	return ctx
}

func provideNoopScopeManager() scope.ScopeManager {
	return &noopScopeManager{}
}
