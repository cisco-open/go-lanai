package internal

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"time"
)

// ClientContextDetails implements
// - security.AuthenticationDetails
// - security.KeyValueDetails
// - oauth2.ClientDetails
// - security.TenantAccessDetails
// It is used to represent a client credential
type ClientContextDetails struct {
	Authentication AuthenticationDetails
	Client         ClientDetails
	KV             map[string]interface{}
	TenantAccess   TenantAccessDetails
}

func (d *ClientContextDetails) ClientId() string {
	return d.Client.Id
}

func (d *ClientContextDetails) AssignedTenantIds() utils.StringSet {
	return d.Client.AssignedTenantIds
}

func (d *ClientContextDetails) Scopes() utils.StringSet {
	return d.Client.Scopes
}

// security.AuthenticationDetails
func (d *ClientContextDetails) ExpiryTime() time.Time {
	return d.Authentication.ExpiryTime
}

// security.AuthenticationDetails
func (d *ClientContextDetails) IssueTime() time.Time {
	return d.Authentication.IssueTime
}

// security.AuthenticationDetails
func (d *ClientContextDetails) Roles() utils.StringSet {
	return d.Authentication.Roles
}

// security.AuthenticationDetails
func (d *ClientContextDetails) Permissions() utils.StringSet {
	return d.Authentication.Permissions
}

// security.AuthenticationDetails
func (d *ClientContextDetails) AuthenticationTime() time.Time {
	return d.Authentication.AuthenticationTime
}

// security.KeyValueDetails
func (d *ClientContextDetails) Value(key string) (v interface{}, ok bool) {
	v, ok = d.KV[key]
	return
}

// security.KeyValueDetails
func (d *ClientContextDetails) Values() (ret map[string]interface{}) {
	ret = map[string]interface{}{}
	for k, v := range d.KV {
		ret[k] = v
	}
	return
}

// ClientTenantedContextDetails implements
// - security.AuthenticationDetails
// - security.KeyValueDetails
// - security.Tenant
// - oauth2.ClientDetails
// - security.TenantAccessDetails
// It is used to represent a client credential with selected tenant
type ClientTenantedContextDetails struct {
	ClientContextDetails
	Tenant TenantDetails
}

func (d *ClientTenantedContextDetails) TenantId() string {
	return d.Tenant.Id
}

func (d *ClientTenantedContextDetails) TenantExternalId() string {
	return d.Tenant.ExternalId
}

func (d *ClientTenantedContextDetails) TenantSuspended() bool {
	return d.Tenant.Suspended
}
