package sectest

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/integrate/security/seclient"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"fmt"
	"time"
)

/*************************
	Mocks
 *************************/

type mockedAuthClient struct {
	*mockedBase
	tokenExp time.Duration
}

// ClientCredentials mocked function will accept any clientID as long as it is accompanied
// by a client Secret.
// If the tokenExp is not defined, will default to 3600
func (c *mockedAuthClient) ClientCredentials(ctx context.Context, opts ...seclient.AuthOptions) (*seclient.Result, error) {
	opt, err := c.option(opts)
	if err != nil {
		return nil, err
	}
	if opt.ClientID != "" || opt.ClientSecret != "" {
		return nil, fmt.Errorf("clientID and clientSecret need to be defined")
	}
	tokenExp := c.tokenExp
	if tokenExp == 0 {
		tokenExp = 3600
	}
	now := time.Now()
	exp := now.UTC().Add(tokenExp)
	return &seclient.Result{
		Request: nil,
		Token: &MockedToken{
			ExpTime:      exp,
			IssTime:      now,
			MockedScopes: opt.Scopes,
		},
	}, nil

}

func newMockedAuthClient(props *mockingProperties, base *mockedBase) seclient.AuthenticationClient {
	return &mockedAuthClient{
		mockedBase: base,
		tokenExp:   time.Duration(props.TokenValidity),
	}
}

func (c *mockedAuthClient) PasswordLogin(_ context.Context, opts ...seclient.AuthOptions) (*seclient.Result, error) {
	opt, e := c.option(opts)
	if e != nil {
		return nil, e
	}
	if opt.AccessToken != "" {
		return nil, fmt.Errorf("[Mocked Error] access token is not allowed for password login")
	}

	acct := c.accounts.find(opt.Username, "")
	if acct == nil || acct.Password != opt.Password {
		return nil, fmt.Errorf("[Mocked Error] username and password don't match")
	}

	tenant, e := c.resolveTenant(opt, acct)
	if e != nil {
		return nil, e
	}

	exp := time.Now().UTC().Add(c.tokenExp)
	return &seclient.Result{
		Token: c.newMockedToken(acct, tenant, exp, ""),
	}, nil
}

func (c *mockedAuthClient) SwitchUser(_ context.Context, opts ...seclient.AuthOptions) (*seclient.Result, error) {
	opt, e := c.option(opts)
	if e != nil {
		return nil, e
	}

	mt, e := c.parseMockedToken(opt.AccessToken)
	if e != nil || mt.UName == "" {
		return nil, fmt.Errorf("[Mocked Error] invalid access token")
	}

	if acct := c.accounts.find(mt.UName, mt.UID); acct == nil || !acct.MockedAccountDetails.Permissions.Has(security.SpecialPermissionSwitchUser) {
		return nil, fmt.Errorf("[Mocked Error] switch user not allowed")
	}

	acct := c.accounts.find(opt.Username, opt.UserId)
	if acct == nil {
		return nil, fmt.Errorf("[Mocked Error] target user doesn't exists")
	}

	tenant, e := c.resolveTenant(opt, acct)
	if e != nil {
		return nil, e
	}

	exp := time.Now().UTC().Add(c.tokenExp)
	return &seclient.Result{
		Token: c.newMockedToken(acct, tenant, exp, mt.UName),
	}, nil
}

func (c *mockedAuthClient) SwitchTenant(_ context.Context, opts ...seclient.AuthOptions) (*seclient.Result, error) {
	opt, e := c.option(opts)
	if e != nil {
		return nil, e
	}

	if opt.Username != "" || opt.UserId != "" {
		return nil, fmt.Errorf("[Mocked Error] username or userId not allowed in switching tenant")
	}

	mt, e := c.parseMockedToken(opt.AccessToken)
	if e != nil || mt.UName == "" {
		return nil, fmt.Errorf("[Mocked Error] invalid access token")
	}

	acct := c.accounts.find(mt.UName, mt.UID)
	if acct == nil || !acct.MockedAccountDetails.Permissions.Has(security.SpecialPermissionSwitchTenant) {
		return nil, fmt.Errorf("[Mocked Error] switch tenant not allowed or deleted user")
	}

	tenant, e := c.resolveTenant(opt, acct)
	if e != nil {
		return nil, e
	}

	exp := time.Now().UTC().Add(c.tokenExp)
	return &seclient.Result{
		Token: c.newMockedToken(acct, tenant, exp, ""),
	}, nil
}

func (c *mockedAuthClient) option(opts []seclient.AuthOptions) (*seclient.AuthOption, error) {
	opt := seclient.AuthOption{}
	for _, fn := range opts {
		fn(&opt)
	}
	if opt.UserId != "" && opt.Username != "" {
		return nil, fmt.Errorf("[Mocked Error] username and userId are exclusive")
	}
	if opt.TenantId != "" && opt.TenantExternalId != "" {
		return nil, fmt.Errorf("[Mocked Error] username and userId are exclusive")
	}
	return &opt, nil
}

func (c *mockedAuthClient) resolveTenant(opt *seclient.AuthOption, acct *MockedAccount) (ret *mockedTenant, err error) {
	if opt.TenantId != "" || opt.TenantExternalId != "" {
		ret = c.tenants.find(opt.TenantId, opt.TenantExternalId)
	} else if acct.DefaultTenant != "" {
		ret = c.tenants.find(acct.DefaultTenant, "")
	}

	if ret == nil {
		return nil, fmt.Errorf("[Mocked Error] tenant not specified and default tenant not configured")
	}

	if !acct.AssignedTenants.Has(ret.ID) && !acct.AssignedTenants.Has(security.SpecialTenantIdWildcard) {
		return nil, fmt.Errorf("[Mocked Error] user does not have access to tenant [%s]", ret.ID)
	}
	return
}
