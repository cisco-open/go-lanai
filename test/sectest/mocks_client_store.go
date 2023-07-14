package sectest

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"fmt"
	"time"
)

var (
	defaultClientGrantTypes = utils.NewStringSet(
		oauth2.GrantTypeClientCredentials,
		oauth2.GrantTypePassword,
		oauth2.GrantTypeAuthCode,
		oauth2.GrantTypeImplicit,
		oauth2.GrantTypeRefresh,
		oauth2.GrantTypeSwitchUser,
		oauth2.GrantTypeSwitchTenant,
		oauth2.GrantTypeSamlSSO,
	)

	defaultClientScopes = utils.NewStringSet(
		oauth2.ScopeRead, oauth2.ScopeWrite,
		oauth2.ScopeTokenDetails, oauth2.ScopeTenantHierarchy,
		oauth2.ScopeOidc, oauth2.ScopeOidcProfile, oauth2.ScopeOidcEmail,
		oauth2.ScopeOidcAddress, oauth2.ScopeOidcPhone,
	)
)

type MockedClient struct {
	MockedClientProperties
}

func (m MockedClient) ID() interface{} {
	return m.MockedClientProperties.ClientID
}

func (m MockedClient) Type() security.AccountType {
	return security.AccountTypeDefault
}

func (m MockedClient) Username() string {
	return m.MockedClientProperties.ClientID
}

func (m MockedClient) Credentials() interface{} {
	return m.MockedClientProperties.Secret
}

func (m MockedClient) Permissions() []string {
	return nil
}

func (m MockedClient) Disabled() bool {
	return false
}

func (m MockedClient) Locked() bool {
	return false
}

func (m MockedClient) UseMFA() bool {
	return false
}

func (m MockedClient) CacheableCopy() security.Account {
	cp := MockedClient{
		m.MockedClientProperties,
	}
	cp.MockedClientProperties.Secret = ""
	return cp
}

func (m MockedClient) ClientId() string {
	return m.MockedClientProperties.ClientID
}

func (m MockedClient) SecretRequired() bool {
	return len(m.MockedClientProperties.Secret) != 0
}

func (m MockedClient) Secret() string {
	return m.MockedClientProperties.Secret
}

func (m MockedClient) GrantTypes() utils.StringSet {
	if m.MockedClientProperties.GrantTypes == nil {
		return defaultClientGrantTypes
	}
	return utils.NewStringSet(m.MockedClientProperties.GrantTypes...)
}

func (m MockedClient) RedirectUris() utils.StringSet {
	return utils.NewStringSet(m.MockedClientProperties.RedirectUris...)
}

func (m MockedClient) Scopes() utils.StringSet {
	if m.MockedClientProperties.Scopes == nil {
		return defaultClientScopes
	}
	return utils.NewStringSet(m.MockedClientProperties.Scopes...)
}

func (m MockedClient) AutoApproveScopes() utils.StringSet {
	return m.Scopes()
}

func (m MockedClient) AccessTokenValidity() time.Duration {
	return time.Duration(m.MockedClientProperties.ATValidity)
}

func (m MockedClient) RefreshTokenValidity() time.Duration {
	return time.Duration(m.MockedClientProperties.RTValidity)
}

func (m MockedClient) UseSessionTimeout() bool {
	return true
}

func (m MockedClient) TenantRestrictions() utils.StringSet {
	return utils.NewStringSet()
}

func (m MockedClient) ResourceIDs() utils.StringSet {
	return utils.NewStringSet()
}

type MockedClientStore struct {
	idLookup map[string]*MockedClient
}

func NewMockedClientStore(props ...*MockedClientProperties) *MockedClientStore {
	ret := MockedClientStore{
		idLookup: map[string]*MockedClient{},
	}
	for _, v := range props {
		ret.idLookup[v.ClientID] = &MockedClient{MockedClientProperties: *v}
	}
	return &ret
}

func (s *MockedClientStore) LoadClientByClientId(_ context.Context, clientId string) (oauth2.OAuth2Client, error) {
	if c, ok := s.idLookup[clientId]; ok {
		return c, nil
	}
	return nil, fmt.Errorf("cannot find client with client ID [%s]", clientId)
}
