package internal_test

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/integrate/security/seclient"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"github.com/pkg/errors"
	"go.uber.org/fx"
	"time"
)

const (
	PropertiesPrefix = "mocking"
)

/*************************
	Mocks
 *************************/

type mockingProperties struct {
	Accounts      map[string]*mockedAccountProperties `json:"accounts"`
	Tenants       map[string]*mockedTenantProperties  `json:"tenants"`
	TokenValidity utils.Duration                      `json:"token-validity"`
}

type mockedAccountProperties struct {
	UserId        string   `json:"id"` // optional field
	Username      string   `json:"username"`
	Password      string   `json:"password"`
	DefaultTenant string   `json:"default-tenant"`
	Tenants       []string `json:"tenants"`
	Perms         []string `json:"permissions"`
}

type mockedTenantProperties struct {
	ID   string `json:"id"` // optional field
	Name string `json:"name"`
}

type MocksDIOut struct {
	fx.Out
	AuthClient   seclient.AuthenticationClient
	TokenReader  oauth2.TokenStoreReader
	TokenRevoker MockedTokenRevoker
	Counter      InvocationCounter
}

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

func ProvideScopeMocksWithCounter(ctx *bootstrap.ApplicationContext) MocksDIOut {
	props := bindMockingProperties(ctx)
	accounts := newMockedAccounts(props)
	tenants := newMockedTenants(props)
	base := mockedBase{
		accounts: accounts,
		tenants:  tenants,
		revoked:  utils.NewStringSet(),
	}
	client := newMockedAuthClient(props, &base)
	reader := newMockedTokenStoreReader(&base)
	counter := invocationCounter{
		AuthenticationClient: client,
		TokenStoreReader:     reader,
		counts:               map[interface{}]*uint64{},
	}
	return MocksDIOut{
		AuthClient:   &counter,
		TokenReader:  &counter,
		TokenRevoker: &base,
		Counter:      &counter,
	}
}

func bindMockingProperties(ctx *bootstrap.ApplicationContext) *mockingProperties {
	props := mockingProperties{
		Accounts:      map[string]*mockedAccountProperties{},
		Tenants:       map[string]*mockedTenantProperties{},
		TokenValidity: utils.Duration(120 * time.Second),
	}
	if err := ctx.Config().Bind(&props, PropertiesPrefix); err != nil {
		panic(errors.Wrap(err, "failed to bind mocking properties"))
	}
	return &props
}

func newMockedAuthClient(props *mockingProperties, base *mockedBase) seclient.AuthenticationClient {
	return &mockedAuthClient{
		mockedBase: base,
		tokenExp:   time.Duration(props.TokenValidity),
	}
}

func newMockedTokenStoreReader(base *mockedBase) oauth2.TokenStoreReader {
	return &mockedTokenStoreReader{
		mockedBase: base,
	}
}
