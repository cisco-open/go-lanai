package testdata

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/sectest"
)

/*************************
	Mocked Security
 *************************/

const (
	OwnerUserId     = "20523d89-d5e9-40d0-afe3-3a74c298b55a"
	AnotherUserId   = "e7498b90-cec3-41fd-ac20-acd41769fb88"
	RootTenantId    = "7b3934fc-edc4-4a1c-9249-3dc7055eb124"
	TenantId        = "8eebb711-7d24-48fb-94da-361c573d7c20"
	AnotherTenantId = "b11ef279-1309-4c43-8355-99c9d494097b"
	ProviderId      = "fe3ad89c-449f-42f2-b4f8-b10ab7bc0266"
)

func MemberAdminOptions() sectest.SecurityContextOptions {
	return sectest.MockedAuthentication(func(d *sectest.SecurityDetailsMock) {
		d.Username = "admin"
		d.UserId = AnotherUserId
		d.TenantId = TenantId
		d.ProviderId = ProviderId
		d.Permissions = utils.NewStringSet("IS_API_ADMIN", "VIEW", "MANAGE")
		d.Roles = utils.NewStringSet("SUPERUSER")
		d.Tenants = utils.NewStringSet(TenantId, AnotherTenantId)
	})
}

func MemberOwnerOptions() sectest.SecurityContextOptions {
	return sectest.MockedAuthentication(func(d *sectest.SecurityDetailsMock) {
		d.Username = "testuser-member-owner"
		d.UserId = OwnerUserId
		d.TenantId = TenantId
		d.ProviderId = ProviderId
		d.Permissions = utils.NewStringSet("VIEW")
		d.Roles = utils.NewStringSet("USER")
		d.Tenants = utils.NewStringSet(TenantId)
	})
}

func MemberNonOwnerOptions() sectest.SecurityContextOptions {
	return sectest.MockedAuthentication(func(d *sectest.SecurityDetailsMock) {
		d.Username = "testuser-member-non-owner"
		d.UserId = AnotherUserId
		d.TenantId = TenantId
		d.ProviderId = ProviderId
		d.Permissions = utils.NewStringSet("VIEW")
		d.Roles = utils.NewStringSet("USER")
		d.Tenants = utils.NewStringSet(TenantId)
	})
}

func NonMemberAdminOptions() sectest.SecurityContextOptions {
	return sectest.MockedAuthentication(func(d *sectest.SecurityDetailsMock) {
		d.Username = "testuser-non-member"
		d.UserId = AnotherUserId
		d.TenantId = AnotherTenantId
		d.ProviderId = ProviderId
		d.Permissions = utils.NewStringSet("IS_API_ADMIN")
		d.Roles = utils.NewStringSet("SUPERUSER")
		d.Tenants = utils.NewStringSet(AnotherTenantId)
	})
}
