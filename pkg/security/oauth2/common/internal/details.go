package internal

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"time"
)

type ProviderDetails struct {
	Id string
	Name string
	DisplayName string
}

type TenantDetails struct {
	Id string
	Name string
	Suspended bool
}

type UserDetails struct {
	Id string
	Username string
	AccountType security.AccountType
	AssignedTenantIds utils.StringSet
	LocaleCode string
	CurrencyCode string
	FirstName string
	LastName string
	Email string
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

type ContextDetails struct {
	Provider       ProviderDetails
	Tenant         TenantDetails
	User           UserDetails
	Authentication AuthenticationDetails
	KV             map[string]interface{}
}

// security.ProviderDetails
func (d *ContextDetails) ProviderId() string {
	return d.Provider.Id
}

// security.ProviderDetails
func (d *ContextDetails) ProviderName() string {
	return d.Provider.Name
}

// security.ProviderDetails
func (d *ContextDetails) ProviderDisplayName() string {
	return d.Provider.DisplayName
}

// security.TenantDetails
func (d *ContextDetails) TenantId() string {
	return d.Tenant.Id
}

// security.TenantDetails
func (d *ContextDetails) TenantName() string {
	return d.Tenant.Name
}

// security.TenantDetails
func (d *ContextDetails) TenantSuspended() bool {
	return d.Tenant.Suspended
}

// security.UserDetails
func (d *ContextDetails) UserId() string {
	return d.User.Id
}

// security.UserDetails
func (d *ContextDetails) Username() string {
	return d.User.Username
}

// security.UserDetails
func (d *ContextDetails) AccountType() security.AccountType {
	return d.User.AccountType
}

// security.UserDetails
func (d *ContextDetails) AssignedTenantIds() utils.StringSet {
	return d.User.AssignedTenantIds
}

// security.UserDetails
func (d *ContextDetails) LocaleCode() string {
	return d.User.LocaleCode
}

// security.UserDetails
func (d *ContextDetails) CurrencyCode() string {
	return d.User.CurrencyCode
}

// security.UserDetails
func (d *ContextDetails) FirstName() string {
	return d.User.FirstName
}

// security.UserDetails
func (d *ContextDetails) LastName() string {
	return d.User.LastName
}

// security.UserDetails
func (d *ContextDetails) Email() string {
	return d.User.Email
}

// security.CredentialDetails
func (d *ContextDetails) ExpiryTime() time.Time {
	return d.Authentication.ExpiryTime
}

// security.CredentialDetails
func (d *ContextDetails) IssueTime() time.Time {
	return d.Authentication.IssueTime
}

// security.CredentialDetails
func (d *ContextDetails) Roles() utils.StringSet {
	return d.Authentication.Roles
}

// security.CredentialDetails
func (d *ContextDetails) Permissions() utils.StringSet {
	return d.Authentication.Permissions
}

// security.CredentialDetails
func (d *ContextDetails) AuthenticationTime() time.Time {
	return d.Authentication.AuthenticationTime
}

// security.CredentialDetails
func (d *ContextDetails) OriginalUsername() string {
	return d.Authentication.OriginalUsername
}

// security.CredentialDetails
func (d *ContextDetails) Proxied() bool {
	return d.Authentication.Proxied
}

// security.KeyValueDetails
func (d *ContextDetails) Value(key string) (v interface{}, ok bool) {
	v, ok = d.KV[key]
	return
}

func (d *ContextDetails) SetValue(key string, value interface{}) {
	if value == nil {
		delete(d.KV, key)
	} else {
		d.KV[key] = value
	}
}

func (d *ContextDetails) SetValues(values map[string]interface{}) {
	d.KV = values
}
