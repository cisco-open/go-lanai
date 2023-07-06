package testdata

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/sectest"
)

const MockingPrefix = "mocking"

type MockingProperties struct {
	Accounts    map[string]*sectest.MockedAccountProperties       `json:"accounts"`
	FedAccounts map[string]*sectest.MockedFederatedUserProperties `json:"fed-accounts"`
	Tenants     map[string]*sectest.MockedTenantProperties        `json:"tenants"`
	Clients     map[string]*sectest.MockedClientProperties        `json:"clients"`
}

func BindMockingProperties(appCtx *bootstrap.ApplicationContext) MockingProperties {
	props := MockingProperties{
		Accounts: map[string]*sectest.MockedAccountProperties{},
		Tenants:  map[string]*sectest.MockedTenantProperties{},
	}
	if e := appCtx.Config().Bind(&props, MockingPrefix); e != nil {
		panic(e)
	}
	return props
}

func MapValues[T any](m map[string]T) []T {
	ret := make([]T, 0, len(m))
	for _, v := range m {
		ret = append(ret, v)
	}
	return ret
}
