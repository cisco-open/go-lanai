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

func NewSimpleContextDetails() *SimpleContextDetails {
	return &SimpleContextDetails{
		Authentication: AuthenticationDetails{
			Roles:       utils.NewStringSet(),
			Permissions: utils.NewStringSet(),
		},
		KV: map[string]interface{}{},
	}
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
