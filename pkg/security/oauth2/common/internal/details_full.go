package internal

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"time"
)

type ProviderDetails struct {
	Id               string
	Name             string
	DisplayName      string
	Description      string
	NotificationType string
	Email            string
}

type TenantDetails struct {
	Id         string
	ExternalId string
	Suspended  bool
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

type ClientDetails struct {
	Id                string
	AssignedTenantIds utils.StringSet
	Scopes            utils.StringSet
}

type TenantAccessDetails struct {
	EffectiveAssignedTenantIds utils.StringSet
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

// ClientUserContextDetails implements
// - security.UserDetails
// - security.AuthenticationDetails
// - security.ProxiedUserDetails
// - security.KeyValueDetails
// - oauth2.ClientDetails

type ClientUserContextDetails struct {
	User           UserDetails
	Client         ClientDetails
	TenantAccess   TenantAccessDetails
	Authentication AuthenticationDetails
	KV             map[string]interface{}
}

func (d *ClientUserContextDetails) ClientId() string {
	return d.Client.Id
}

func (d *ClientUserContextDetails) Scopes() utils.StringSet {
	return d.Client.Scopes
}

func NewClientUserContextDetails() *ClientUserContextDetails {
	return &ClientUserContextDetails{
		Client: ClientDetails{
			AssignedTenantIds: utils.NewStringSet(),
			Scopes:            utils.NewStringSet(),
		},
		User: UserDetails{
			AssignedTenantIds: utils.NewStringSet(),
		},
		Authentication: AuthenticationDetails{
			Roles:       utils.NewStringSet(),
			Permissions: utils.NewStringSet(),
		},
		KV: map[string]interface{}{},
		TenantAccess: TenantAccessDetails{
			EffectiveAssignedTenantIds: utils.NewStringSet(),
		},
	}
}

// security.UserDetails
func (d *ClientUserContextDetails) UserId() string {
	return d.User.Id
}

// security.UserDetails
func (d *ClientUserContextDetails) Username() string {
	return d.User.Username
}

// security.UserDetails
func (d *ClientUserContextDetails) AccountType() security.AccountType {
	return d.User.AccountType
}

// security.UserDetails
// Deprecate: the interface is deprecated
func (d *ClientUserContextDetails) AssignedTenantIds() utils.StringSet {
	return d.User.AssignedTenantIds
}

// security.UserDetails
func (d *ClientUserContextDetails) LocaleCode() string {
	return d.User.LocaleCode
}

// security.UserDetails
func (d *ClientUserContextDetails) CurrencyCode() string {
	return d.User.CurrencyCode
}

// security.UserDetails
func (d *ClientUserContextDetails) FirstName() string {
	return d.User.FirstName
}

// security.UserDetails
func (d *ClientUserContextDetails) LastName() string {
	return d.User.LastName
}

// security.UserDetails
func (d *ClientUserContextDetails) Email() string {
	return d.User.Email
}

// security.AuthenticationDetails
func (d *ClientUserContextDetails) ExpiryTime() time.Time {
	return d.Authentication.ExpiryTime
}

// security.AuthenticationDetails
func (d *ClientUserContextDetails) IssueTime() time.Time {
	return d.Authentication.IssueTime
}

// security.AuthenticationDetails
func (d *ClientUserContextDetails) Roles() utils.StringSet {
	return d.Authentication.Roles
}

// security.AuthenticationDetails
func (d *ClientUserContextDetails) Permissions() utils.StringSet {
	return d.Authentication.Permissions
}

// security.AuthenticationDetails
func (d *ClientUserContextDetails) AuthenticationTime() time.Time {
	return d.Authentication.AuthenticationTime
}

// security.ProxiedUserDetails
func (d *ClientUserContextDetails) OriginalUsername() string {
	return d.Authentication.OriginalUsername
}

// security.ProxiedUserDetails
func (d *ClientUserContextDetails) Proxied() bool {
	return d.Authentication.Proxied
}

// security.KeyValueDetails
func (d *ClientUserContextDetails) Value(key string) (v interface{}, ok bool) {
	v, ok = d.KV[key]
	return
}

// security.KeyValueDetails
func (d *ClientUserContextDetails) Values() (ret map[string]interface{}) {
	ret = map[string]interface{}{}
	for k, v := range d.KV {
		ret[k] = v
	}
	return
}

func (d *ClientUserContextDetails) EffectiveAssignedTenantIds() utils.StringSet {
	return d.TenantAccess.EffectiveAssignedTenantIds
}

// ClientUserContextDetails implements
// - security.UserDetails
// - security.TenantDetails
// - security.ProviderDetails
// - security.AuthenticationDetails
// - security.ProxiedUserDetails
// - security.KeyValueDetails
// - oauth2.ClientDetails

type ClientUserTenantedContextDetails struct {
	ClientUserContextDetails
	Tenant   TenantDetails
	Provider ProviderDetails
}

func NewClientUserTenantedContextDetails() *ClientUserTenantedContextDetails {
	return &ClientUserTenantedContextDetails{
		ClientUserContextDetails: ClientUserContextDetails{
			Client: ClientDetails{
				AssignedTenantIds: utils.NewStringSet(),
				Scopes:            utils.NewStringSet(),
			},
			User: UserDetails{
				AssignedTenantIds: utils.NewStringSet(),
			},
			Authentication: AuthenticationDetails{
				Roles:       utils.NewStringSet(),
				Permissions: utils.NewStringSet(),
			},
			KV: map[string]interface{}{},
			TenantAccess: TenantAccessDetails{
				EffectiveAssignedTenantIds: utils.NewStringSet(),
			},
		},
		Tenant:   TenantDetails{},
		Provider: ProviderDetails{},
	}
}

// security.TenantDetails
func (d *ClientUserTenantedContextDetails) TenantId() string {
	return d.Tenant.Id
}

// security.TenantDetails
func (d *ClientUserTenantedContextDetails) TenantExternalId() string {
	return d.Tenant.ExternalId
}

// security.TenantDetails
func (d *ClientUserTenantedContextDetails) TenantSuspended() bool {
	return d.Tenant.Suspended
}

// security.ProviderDetails
func (d *ClientUserTenantedContextDetails) ProviderId() string {
	return d.Provider.Id
}

// security.ProviderDetails
func (d *ClientUserTenantedContextDetails) ProviderName() string {
	return d.Provider.Name
}

// security.ProviderDetails
func (d *ClientUserTenantedContextDetails) ProviderDisplayName() string {
	return d.Provider.DisplayName
}

func (d *ClientUserTenantedContextDetails) ProviderDescription() string {
	return d.Provider.Description
}

func (d *ClientUserTenantedContextDetails) ProviderEmail() string {
	return d.Provider.Email
}

func (d *ClientUserTenantedContextDetails) ProviderNotificationType() string {
	return d.Provider.NotificationType
}
