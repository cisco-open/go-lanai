package sectest

import (
	appconfig "cto-github.cisco.com/NFV-BU/go-lanai/pkg/appconfig/init"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	securityint "cto-github.cisco.com/NFV-BU/go-lanai/pkg/integrate/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/integrate/security/scope"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/integrate/security/seclient"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/apptest"
	"embed"
	"go.uber.org/fx"
)

//var logger = log.New("SEC.Test")

//go:embed test-scopes.yml
var defaultMockingConfigFS embed.FS

/**************************
	Options
 **************************/

// WithMockedScopes is a test.Options that initialize cto-github.cisco.com/NFV-BU/go-lanai/pkg/integrate/security/scope
// This option configure mocked security scopes based on yaml provided as embed.FS.
// If no config is provided, the default config is used
func WithMockedScopes(mocksConfigFS ...embed.FS) test.Options {
	fxOpts := make([]fx.Option, len(mocksConfigFS), len(mocksConfigFS) + 3)
	for i, fs := range mocksConfigFS {
		fxOpts[i] = appconfig.FxEmbeddedApplicationAdHoc(fs)
	}
	fxOpts = append(fxOpts,
		appconfig.FxEmbeddedBootstrapAdHoc(defaultMockingConfigFS),
		fx.Provide(securityint.BindSecurityIntegrationProperties),
		fx.Provide(ProvideScopeMocks),
	)
	opts := []test.Options{
		apptest.WithModules(scope.Module),
		apptest.WithFxOptions(fxOpts...),
	}
	return func(opt *test.T) {
		for _, fn := range opts {
			fn(opt)
		}
	}
}

/**************************
	fx options
 **************************/

type MocksDIOut struct {
	fx.Out
	AuthClient   seclient.AuthenticationClient
	TokenReader  oauth2.TokenStoreReader
	TokenRevoker MockedTokenRevoker
}

// ProvideScopeMocks is for internal usage. Exported for cross-package reference
// Try use WithMockedScopes instead
func ProvideScopeMocks(ctx *bootstrap.ApplicationContext) MocksDIOut {
	props := bindMockingProperties(ctx)
	accounts := newMockedAccounts(props)
	tenants := newMockedTenants(props)
	base := mockedBase{
		accounts: accounts,
		tenants:  tenants,
		revoked:  utils.NewStringSet(),
	}

	return MocksDIOut{
		AuthClient:   newMockedAuthClient(props, &base),
		TokenReader:  newMockedTokenStoreReader(&base),
		TokenRevoker: &base,
	}
}
