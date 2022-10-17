package security

import "cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"

type MockedAccountAuth struct {
	permissions Permissions
	details     UserDetails
}

type MockedUserDetails struct {
	userId            string
	username          string
	assignedTenantIds utils.StringSet
}

func (m MockedUserDetails) UserId() string {
	return m.userId
}

func (m MockedUserDetails) Username() string {
	return m.username
}

func (m MockedUserDetails) AccountType() AccountType {
	return 0
}

func (m MockedUserDetails) AssignedTenantIds() utils.StringSet {
	return m.assignedTenantIds
}

func (m MockedUserDetails) LocaleCode() string {
	return ""
}

func (m MockedUserDetails) CurrencyCode() string {
	return ""
}

func (m MockedUserDetails) FirstName() string {
	return ""
}

func (m MockedUserDetails) LastName() string {
	return ""
}

func (m MockedUserDetails) Email() string {
	return ""
}

func (m MockedAccountAuth) Principal() interface{} {
	return nil
}

func (m MockedAccountAuth) Permissions() Permissions {
	perms := Permissions{}
	for perm := range m.permissions {
		perms[perm] = struct{}{}
	}
	return perms
}

func (m MockedAccountAuth) State() AuthenticationState {
	return StateAuthenticated
}

func (m MockedAccountAuth) Details() interface{} {
	return m.details
}
