package sectest

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"time"
)

type SecurityMockOptions func(d *SecurityDetailsMock)

type SecurityDetailsMock struct {
	Username                 string
	UserId                   string
	TenantExternalId         string
	TenantId                 string
	ProviderName             string
	ProviderId               string
	ProviderDisplayName      string
	ProviderDescription      string
	ProviderEmail            string
	ProviderNotificationType string
	AccessToken              string
	Exp                      time.Time
	Iss                      time.Time
	Permissions              utils.StringSet
	Tenants                  utils.StringSet
	OrigUsername             string
	UserFirstName            string
	UserLastName             string
	KVs                      map[string]interface{}
	ClientID                 string
	Scopes                   utils.StringSet
	UserEmail                string
}

// MockedSecurityDetails implements
// - security.AuthenticationDetails
// - security.ProxiedUserDetails
// - security.UserDetails
// - security.TenantDetails
// - security.ProviderDetails
// - security.KeyValueDetails
type MockedSecurityDetails struct {
	SecurityDetailsMock
}

func NewMockedSecurityDetails(opts ...SecurityMockOptions) *MockedSecurityDetails {
	ret := MockedSecurityDetails{
		SecurityDetailsMock{
			ClientID: "mock",
		},
	}
	for _, fn := range opts {
		fn(&ret.SecurityDetailsMock)
	}
	return &ret
}

func (d *MockedSecurityDetails) Value(s string) (interface{}, bool) {
	v, ok := d.KVs[s]
	return v, ok
}

func (d *MockedSecurityDetails) Values() map[string]interface{} {
	return d.KVs
}

func (d *MockedSecurityDetails) OriginalUsername() string {
	return d.OrigUsername
}

func (d *MockedSecurityDetails) Proxied() bool {
	return d.OrigUsername != ""
}

func (d *MockedSecurityDetails) ExpiryTime() time.Time {
	return d.Exp
}

func (d *MockedSecurityDetails) IssueTime() time.Time {
	return d.Iss
}

func (d *MockedSecurityDetails) Roles() utils.StringSet {
	panic("implement me")
}

func (d *MockedSecurityDetails) Permissions() utils.StringSet {
	if d.SecurityDetailsMock.Permissions == nil {
		d.SecurityDetailsMock.Permissions = utils.NewStringSet()
	}
	return d.SecurityDetailsMock.Permissions
}

func (d *MockedSecurityDetails) AuthenticationTime() time.Time {
	return d.Iss
}

func (d *MockedSecurityDetails) ProviderId() string {
	return d.SecurityDetailsMock.ProviderId
}

func (d *MockedSecurityDetails) ProviderName() string {
	return d.SecurityDetailsMock.ProviderName
}

func (d *MockedSecurityDetails) ProviderDisplayName() string {
	return d.SecurityDetailsMock.ProviderDisplayName
}

func (d *MockedSecurityDetails) ProviderDescription() string {
	return d.SecurityDetailsMock.ProviderDescription
}

func (d *MockedSecurityDetails) ProviderEmail() string {
	return d.SecurityDetailsMock.ProviderEmail
}

func (d *MockedSecurityDetails) ProviderNotificationType() string {
	return d.SecurityDetailsMock.ProviderNotificationType
}

func (d *MockedSecurityDetails) TenantId() string {
	return d.SecurityDetailsMock.TenantId
}

func (d *MockedSecurityDetails) TenantExternalId() string {
	return d.SecurityDetailsMock.TenantExternalId
}

func (d *MockedSecurityDetails) TenantSuspended() bool {
	return false
}

func (d *MockedSecurityDetails) UserId() string {
	return d.SecurityDetailsMock.UserId
}

func (d *MockedSecurityDetails) Username() string {
	return d.SecurityDetailsMock.Username
}

func (d *MockedSecurityDetails) AccountType() security.AccountType {
	return security.AccountTypeDefault
}

func (d *MockedSecurityDetails) AssignedTenantIds() utils.StringSet {
	if d.Tenants == nil {
		d.Tenants = utils.NewStringSet()
	}
	return d.Tenants
}

func (d *MockedSecurityDetails) LocaleCode() string {
	return ""
}

func (d *MockedSecurityDetails) CurrencyCode() string {
	return ""
}

func (d *MockedSecurityDetails) FirstName() string {
	return d.UserFirstName
}

func (d *MockedSecurityDetails) LastName() string {
	return d.UserLastName
}

func (d *MockedSecurityDetails) Email() string {
	return d.UserEmail
}
