package testdata

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/tenancy"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/mocks"
	"embed"
	"github.com/google/uuid"
)

//go:embed create_table_a.sql model_a.yml
var ModelADataFS embed.FS

var (
	MockedAdminId = uuid.MustParse("710e8219-ed8d-474e-8f7d-96b27e46dba9")
	MockedUserId1 = uuid.MustParse("595959e4-8803-4ab1-8acf-acfb92bb7322")
	MockedUserId2 = uuid.MustParse("9a901c91-a3d6-4d39-9adf-34e74bb32de2")
	MockedRootTenantId = uuid.MustParse("23967dfe-d90f-4e1b-9406-e2df6685f232")
	MockedTenantIdA    = uuid.MustParse("d8423acc-28cb-4209-95d6-089de7fb27ef")
	MockedTenantIdB    = uuid.MustParse("37b7181a-0892-4706-8f26-60d286b63f14")
	MockedTenantIdA1   = uuid.MustParse("be91531e-ca96-46eb-aea6-b7e0e2a50e21")
	MockedTenantIdA2   = uuid.MustParse("b50c18d9-1741-49bd-8536-30943dfffb45")
	MockedTenantIdB1   = uuid.MustParse("1513b015-6a7d-4de3-8b4f-cbb090ac126d")
	MockedTenantIdB2   = uuid.MustParse("b21445de-9192-45de-acd7-91745ab4cc13")
)

/*************************
	Mocks
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