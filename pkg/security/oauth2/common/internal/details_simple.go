package internal

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"time"
)

// SimpleContextDetails implements
// - security.AuthenticationDetails
// - security.KeyValueDetails
type SimpleContextDetails struct {
	Authentication AuthenticationDetails
	KV             map[string]interface{}
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
func (d *SimpleContextDetails) SetValue(key string, value interface{}) {
	if value == nil {
		delete(d.KV, key)
	} else {
		d.KV[key] = value
	}
}

// security.KeyValueDetails
func (d *SimpleContextDetails) SetValues(values map[string]interface{}) {
	d.KV = values
}
