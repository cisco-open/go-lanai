package sectest

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"time"
)

type SecurityMockOptions func(d *SecurityDetailsMock)

type SecurityDetailsMock struct {
	Username     string
	UserId       string
	TenantExternalId   string
	TenantId     string
	ProviderName string
	ProviderId   string
	ProviderDisplayName string
	ProviderDescription string
	ProviderEmail string
	ProviderNotificationType string
	Exp          time.Time
	Iss          time.Time
	Permissions  utils.StringSet
	Tenants      utils.StringSet
	OrigUsername string
	UserFirstName    string
	UserLastName     string
}

// mockedSecurityDetails implements
// - security.AuthenticationDetails
// - security.ProxiedUserDetails
// - security.UserDetails
// - security.TenantDetails
// - security.ProviderDetails
type mockedSecurityDetails struct {
	SecurityDetailsMock
}

func NewMockedSecurityDetails(opts...SecurityMockOptions) security.AuthenticationDetails {
	ret := mockedSecurityDetails{}
	for _, fn := range opts {
		fn(&ret.SecurityDetailsMock)
	}
	return &ret
}

func (d *mockedSecurityDetails) OriginalUsername() string {
	return d.OrigUsername
}

func (d *mockedSecurityDetails) Proxied() bool {
	return d.OrigUsername != ""
}

func (d *mockedSecurityDetails) ExpiryTime() time.Time {
	return d.Exp
}

func (d *mockedSecurityDetails) IssueTime() time.Time {
	return d.Iss
}

func (d *mockedSecurityDetails) Roles() utils.StringSet {
	panic("implement me")
}

func (d *mockedSecurityDetails) Permissions() utils.StringSet {
	if d.SecurityDetailsMock.Permissions == nil {
		d.SecurityDetailsMock.Permissions = utils.NewStringSet()
	}
	return d.SecurityDetailsMock.Permissions
}

func (d *mockedSecurityDetails) AuthenticationTime() time.Time {
	return d.Iss
}

func (d *mockedSecurityDetails) ProviderId() string {
	return d.SecurityDetailsMock.ProviderId
}

func (d *mockedSecurityDetails) ProviderName() string {
	return d.SecurityDetailsMock.ProviderName
}

func (d *mockedSecurityDetails) ProviderDisplayName() string {
	return d.SecurityDetailsMock.ProviderDisplayName
}

func (d *mockedSecurityDetails) ProviderDescription() string {
	return d.SecurityDetailsMock.ProviderDescription
}

func (d *mockedSecurityDetails) ProviderEmail() string {
	return d.SecurityDetailsMock.ProviderEmail
}

func (d *mockedSecurityDetails) ProviderNotificationType() string {
	return d.SecurityDetailsMock.ProviderNotificationType
}


func (d *mockedSecurityDetails) TenantId() string {
	return d.SecurityDetailsMock.TenantId
}

func (d *mockedSecurityDetails) TenantExternalId() string {
	return d.SecurityDetailsMock.TenantExternalId
}

func (d *mockedSecurityDetails) TenantSuspended() bool {
	panic("implement me")
}

func (d *mockedSecurityDetails) UserId() string {
	return d.SecurityDetailsMock.UserId
}

func (d *mockedSecurityDetails) Username() string {
	return d.SecurityDetailsMock.Username
}

func (d *mockedSecurityDetails) AccountType() security.AccountType {
	panic("implement me")
}

func (d *mockedSecurityDetails) AssignedTenantIds() utils.StringSet {
	if d.Tenants == nil {
		d.Tenants = utils.NewStringSet()
	}
	return d.Tenants
}

func (d *mockedSecurityDetails) LocaleCode() string {
	panic("implement me")
}

func (d *mockedSecurityDetails) CurrencyCode() string {
	panic("implement me")
}

func (d *mockedSecurityDetails) FirstName() string {
	return d.UserFirstName
}

func (d *mockedSecurityDetails) LastName() string {
	return d.UserLastName
}

func (d *mockedSecurityDetails) Email() string {
	panic("implement me")
}
