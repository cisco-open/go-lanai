package testdata

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/tenancy"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/mocks"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/sectest"
	"embed"
	"github.com/google/uuid"
)

//go:embed *.sql *.yml
var ModelDataFS embed.FS

var (
	MockedAdminId = uuid.MustParse("710e8219-ed8d-474e-8f7d-96b27e46dba9")
	MockedUserId1 = uuid.MustParse("595959e4-8803-4ab1-8acf-acfb92bb7322")
	MockedUserId2 = uuid.MustParse("9a901c91-a3d6-4d39-9adf-34e74bb32de2")
	MockedUserId3 = uuid.MustParse("e212a869-b636-4dc6-83db-e1ccd59e5e0e")

	MockedRootTenantId = uuid.MustParse("23967dfe-d90f-4e1b-9406-e2df6685f232")
	MockedTenantIdA    = uuid.MustParse("d8423acc-28cb-4209-95d6-089de7fb27ef")
	MockedTenantIdB    = uuid.MustParse("37b7181a-0892-4706-8f26-60d286b63f14")
	MockedTenantIdA1   = uuid.MustParse("be91531e-ca96-46eb-aea6-b7e0e2a50e21")
	MockedTenantIdA2   = uuid.MustParse("b50c18d9-1741-49bd-8536-30943dfffb45")
	MockedTenantIdB1   = uuid.MustParse("1513b015-6a7d-4de3-8b4f-cbb090ac126d")
	MockedTenantIdB2   = uuid.MustParse("b21445de-9192-45de-acd7-91745ab4cc13")
)

/*************************
	Tenancy
 *************************/

func ProvideMockedTenancyAccessor() tenancy.Accessor {
	tenancyRelationship := []mocks.TenancyRelation{
		{Parent: MockedRootTenantId, Child: MockedTenantIdA},
		{Parent: MockedRootTenantId, Child: MockedTenantIdB},
		{Parent: MockedTenantIdA, Child: MockedTenantIdA1},
		{Parent: MockedTenantIdA, Child: MockedTenantIdA2},
		{Parent: MockedTenantIdB, Child: MockedTenantIdB1},
		{Parent: MockedTenantIdB, Child: MockedTenantIdB2},
	}
	return mocks.NewMockTenancyAccessor(tenancyRelationship, MockedRootTenantId)
}

/*************************
	Users
 *************************/

func ContextWithSecurityMock(parent context.Context, mockOpts ...sectest.SecurityMockOptions) context.Context {
	return sectest.ContextWithSecurity(parent, sectest.MockedAuthentication(mockOpts...))
}

func AdminSecurityOptions(tenantId ...uuid.UUID) sectest.SecurityMockOptions {
	return func(d *sectest.SecurityDetailsMock) {
		d.Username = "admin"
		d.UserId = MockedAdminId.String()
		d.TenantExternalId = "Root Tenant"
		d.Permissions = utils.NewStringSet("VIEW", "MANAGE")
		d.Roles = utils.NewStringSet("ADMIN")
		d.Tenants = utils.NewStringSet(MockedRootTenantId.String())
		d.TenantId = MockedRootTenantId.String()
		if len(tenantId) != 0 {
			d.TenantId = tenantId[0].String()
			d.Tenants.Add(d.TenantId)
		}
	}
}

func User1SecurityOptions(tenantId ...uuid.UUID) sectest.SecurityMockOptions {
	return func(d *sectest.SecurityDetailsMock) {
		d.Username = "user1"
		d.UserId = MockedUserId1.String()
		d.TenantExternalId = "Tenant A"
		d.Permissions = utils.NewStringSet("NO_VIEW")
		d.Roles = utils.NewStringSet("USER")
		d.Tenants = utils.NewStringSet(MockedTenantIdA.String())
		d.TenantId = MockedTenantIdA1.String()
		if len(tenantId) != 0 {
			d.TenantId = tenantId[0].String()
			d.Tenants.Add(d.TenantId)
		}
	}
}

func User2SecurityOptions(tenantId ...uuid.UUID) sectest.SecurityMockOptions {
	return func(d *sectest.SecurityDetailsMock) {
		d.Username = "user2"
		d.UserId = MockedUserId2.String()
		d.TenantExternalId = "Tenant B"
		d.Permissions = utils.NewStringSet("NO_VIEW")
		d.Roles = utils.NewStringSet("USER")
		d.Tenants = utils.NewStringSet(MockedTenantIdB.String())
		d.TenantId = MockedTenantIdB1.String()
		if len(tenantId) != 0 {
			d.TenantId = tenantId[0].String()
			d.Tenants.Add(d.TenantId)
		}
	}
}

func User3SecurityOptions(tenantId ...uuid.UUID) sectest.SecurityMockOptions {
	return func(d *sectest.SecurityDetailsMock) {
		d.Username = "user3"
		d.UserId = MockedUserId3.String()
		d.TenantExternalId = "Tenant A"
		d.Permissions = utils.NewStringSet("NO_VIEW")
		d.Roles = utils.NewStringSet("USER")
		d.Tenants = utils.NewStringSet(MockedTenantIdA.String())
		d.TenantId = MockedTenantIdA1.String()
		if len(tenantId) != 0 {
			d.TenantId = tenantId[0].String()
			d.Tenants.Add(d.TenantId)
		}
	}
}

func ExtraPermsSecurityOptions(permissions ...string) sectest.SecurityMockOptions {
	return func(d *sectest.SecurityDetailsMock) {
		d.Permissions.Add(permissions...)
	}
}
