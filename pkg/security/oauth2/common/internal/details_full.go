package internal

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"time"
)

type ProviderDetails struct {
	Id          string
	Name        string
	DisplayName string
}

type TenantDetails struct {
	Id        string
	Name      string
	Suspended bool
}

type UserDetails struct {
	Id                string
	Username          string
	AccountType       security.AccountType
	AssignedTenantIds utils.StringSet
	LocaleCode        string
	CurrencyCode      string
	FirstName         string
	LastName          string
	Email             string
}

type AuthenticationDetails struct {
	IssueTime          time.Time
	ExpiryTime         time.Time
	Roles              utils.StringSet
	Permissions        utils.StringSet
	AuthenticationTime time.Time
	OriginalUsername   string
	Proxied            bool
}

// FullContextDetails implements
// - security.UserDetails
// - security.TenantDetails
// - security.ProviderDetails
// - security.AuthenticationDetails
// - security.ProxiedUserDetails
// - security.KeyValueDetails
type FullContextDetails struct {
	Provider       ProviderDetails
	Tenant         TenantDetails
	User           UserDetails
	Authentication AuthenticationDetails
	KV             map[string]interface{}
}

func NewFullContextDetails() *FullContextDetails {
	return &FullContextDetails{
		Provider: ProviderDetails{},
		Tenant:   TenantDetails{},
		User: UserDetails{
			AssignedTenantIds: utils.NewStringSet(),
		},
		Authentication: AuthenticationDetails{
			Roles:       utils.NewStringSet(),
			Permissions: utils.NewStringSet(),
		},
		KV: map[string]interface{}{},
	}
}

// security.ProviderDetails
func (d *FullContextDetails) ProviderId() string {
	return d.Provider.Id
}

// security.ProviderDetails
func (d *FullContextDetails) ProviderName() string {
	return d.Provider.Name
}

// security.ProviderDetails
func (d *FullContextDetails) ProviderDisplayName() string {
	return d.Provider.DisplayName
}

// security.TenantDetails
func (d *FullContextDetails) TenantId() string {
	return d.Tenant.Id
}

// security.TenantDetails
func (d *FullContextDetails) TenantName() string {
	return d.Tenant.Name
}

// security.TenantDetails
func (d *FullContextDetails) TenantSuspended() bool {
	return d.Tenant.Suspended
}

// security.UserDetails
func (d *FullContextDetails) UserId() string {
	return d.User.Id
}

// security.UserDetails
func (d *FullContextDetails) Username() string {
	return d.User.Username
}

// security.UserDetails
func (d *FullContextDetails) AccountType() security.AccountType {
	return d.User.AccountType
}

// security.UserDetails
func (d *FullContextDetails) AssignedTenantIds() utils.StringSet {
	return d.User.AssignedTenantIds
}

// security.UserDetails
func (d *FullContextDetails) LocaleCode() string {
	return d.User.LocaleCode
}

// security.UserDetails
func (d *FullContextDetails) CurrencyCode() string {
	return d.User.CurrencyCode
}

// security.UserDetails
func (d *FullContextDetails) FirstName() string {
	return d.User.FirstName
}

// security.UserDetails
func (d *FullContextDetails) LastName() string {
	return d.User.LastName
}

// security.UserDetails
func (d *FullContextDetails) Email() string {
	return d.User.Email
}

// security.AuthenticationDetails
func (d *FullContextDetails) ExpiryTime() time.Time {
	return d.Authentication.ExpiryTime
}

// security.AuthenticationDetails
func (d *FullContextDetails) IssueTime() time.Time {
	return d.Authentication.IssueTime
}

// security.AuthenticationDetails
func (d *FullContextDetails) Roles() utils.StringSet {
	return d.Authentication.Roles
}

// security.AuthenticationDetails
func (d *FullContextDetails) Permissions() utils.StringSet {
	return d.Authentication.Permissions
}

// security.AuthenticationDetails
func (d *FullContextDetails) AuthenticationTime() time.Time {
	return d.Authentication.AuthenticationTime
}

// security.ProxiedUserDetails
func (d *FullContextDetails) OriginalUsername() string {
	return d.Authentication.OriginalUsername
}

// security.ProxiedUserDetails
func (d *FullContextDetails) Proxied() bool {
	return d.Authentication.Proxied
}

// security.KeyValueDetails
func (d *FullContextDetails) Value(key string) (v interface{}, ok bool) {
	v, ok = d.KV[key]
	return
}

// security.KeyValueDetails
func (d *FullContextDetails) Values() (ret map[string]interface{}) {
	ret = map[string]interface{}{}
	for k, v := range d.KV {
		ret[k] = v
	}
	return
}



