package internal

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"time"
)

// SimpleContextDetails implements
// - security.AuthenticationDetails
// - security.KeyValueDetails
// It is used to represent a client credential
type SimpleContextDetails struct {
	Authentication AuthenticationDetails
	Tenant         TenantDetails
	Client         ClientDetails
	KV             map[string]interface{}
}

func (d *SimpleContextDetails) ClientId() string {
	return d.Client.Id
}

func (d *SimpleContextDetails) AssignedTenantIds() utils.StringSet {
	return d.Client.AssignedTenantIds
}

func (d *SimpleContextDetails) Scopes() utils.StringSet {
	return d.Client.Scopes
}

func (d *SimpleContextDetails) TenantId() string {
	return d.Tenant.Id
}

func (d *SimpleContextDetails) TenantExternalId() string {
	return d.Tenant.ExternalId
}

func (d *SimpleContextDetails) TenantSuspended() bool {
	return d.Tenant.Suspended
}

// security.AuthenticationDetails
func (d *SimpleContextDetails) ExpiryTime() time.Time {
	return d.Authentication.ExpiryTime
}

// security.AuthenticationDetails
func (d *SimpleContextDetails) IssueTime() time.Time {
	return d.Authentication.IssueTime
}

// security.AuthenticationDetails
func (d *SimpleContextDetails) Roles() utils.StringSet {
	return d.Authentication.Roles
}

// security.AuthenticationDetails
func (d *SimpleContextDetails) Permissions() utils.StringSet {
	return d.Authentication.Permissions
}

// security.AuthenticationDetails
func (d *SimpleContextDetails) AuthenticationTime() time.Time {
	return d.Authentication.AuthenticationTime
}

// security.KeyValueDetails
func (d *SimpleContextDetails) Value(key string) (v interface{}, ok bool) {
	v, ok = d.KV[key]
	return
}

// security.KeyValueDetails
func (d *SimpleContextDetails) Values() (ret map[string]interface{}) {
	ret = map[string]interface{}{}
	for k, v := range d.KV {
		ret[k] = v
	}
	return
}
