package sectest

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"fmt"
	"sync"
	"time"
)

/*************************
	Interface
 *************************/

type MockedTokenRevoker interface {
	Revoke(value string)
	RevokeAll()
}

/*************************
	Base
 *************************/

// mockedBase implements MockedTokenRevoker
type mockedBase struct {
	sync.RWMutex
	accounts  *mockedAccounts
	tenants   *mockedTenants
	revoked   utils.StringSet
	notBefore time.Time
}

func (b *mockedBase) Revoke(value string) {
	b.Lock()
	defer b.Unlock()
	b.revoked.Add(value)
}

func (b *mockedBase) RevokeAll() {
	b.Lock()
	defer b.Unlock()
	b.notBefore = time.Now().UTC()
}

func (b *mockedBase) isTokenRevoked(token *MockedToken, value string) bool {
	b.RLock()
	defer b.RUnlock()
	return token.IssTime.Before(b.notBefore) || b.revoked.Has(value)
}

func (b *mockedBase) newMockedToken(acct *mockedAccount, tenant *mockedTenant, exp time.Time, origUser string) *MockedToken {
	return &MockedToken{
		MockedTokenInfo: MockedTokenInfo{
			UName: acct.Username,
			UID:   acct.UserId,
			TID:   tenant.ID,
			TExternalId: tenant.ExternalId,
			OrigU: origUser,
		},
		ExpTime: exp,
		IssTime: time.Now().UTC(),
	}
}

func (b *mockedBase) parseMockedToken(value string) (*MockedToken, error) {
	mt := &MockedToken{}
	if e := mt.UnmarshalText([]byte(value)); e != nil {
		return nil, e
	}
	if b.isTokenRevoked(mt, value) {
		return nil, fmt.Errorf("[Mocked Error]: token revoked")
	}
	return mt, nil
}

func (b *mockedBase) newMockedAuth(mt *MockedToken, acct *mockedAccount) oauth2.Authentication {
	user := oauth2.NewUserAuthentication(func(opt *oauth2.UserAuthOption) {
		opt.Principal = mt.UName
		opt.State = security.StateAuthenticated
		opt.Permissions = map[string]interface{}{}
		for perm := range acct.Permissions {
			opt.Permissions[perm] = true
		}
	})
	details := NewMockedSecurityDetails(func(d *SecurityDetailsMock) {
		*d = SecurityDetailsMock{
			Username:     mt.UName,
			UserId:       mt.UID,
			TenantExternalId:   mt.TExternalId,
			TenantId:     mt.TID,
			Exp:          mt.ExpTime,
			Iss:          mt.IssTime,
			Permissions:  acct.Permissions,
			Tenants:      acct.AssignedTenants,
			OrigUsername: mt.OrigU,
		}
	})
	return oauth2.NewAuthentication(func(opt *oauth2.AuthOption) {
		opt.Request = oauth2.NewOAuth2Request(func(opt *oauth2.RequestDetails) {
			opt.ClientId = "mock"
			opt.Approved = true
		})
		opt.Token = mt
		opt.UserAuth = user
		opt.Details = details
	})
}
