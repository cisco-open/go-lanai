package opadata

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/data/types/pqx"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/opa"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/apptest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/dbtest"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"testing"
)

/*************************
	Test
 *************************/

func TestResourceResolver(t *testing.T) {
	di := &TestDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		//apptest.WithTimeout(10*time.Minute),
		dbtest.WithNoopMocks(),
		apptest.WithDI(di),
		test.GomegaSubTest(SubTestResolveResource(), "TestResolveResource"),
		test.GomegaSubTest(SubTestResolveInvalidModel(), "TestResolveInvalidModel"),
		test.GomegaSubTest(SubTestInvalidGenerics(), "TestInvalidGenerics"),
	)
}

/*************************
	Sub Tests
 *************************/

func SubTestResolveResource() test.GomegaSubTestFunc {
	type Model struct {
		ID         uuid.UUID     `gorm:"primaryKey;type:uuid;default:gen_random_uuid();"`
		TenantID   string        `gorm:"not null" opa:"field:tenant_id"`
		TenantPath pqx.UUIDArray `gorm:"type:uuid[];index:,type:gin;not null" opa:"field:tenant_path"`
		OwnerID    uuid.UUID     `gorm:"type:KeyID;not null" opa:"field:owner_id"`
		Sharing    pqx.JsonbMap  `gorm:"type:jsonb;not null" opa:"field:sharing"`
		OPAArray   []string      `gorm:"type:text[]" opa:"field:extra_array"`
		OPASingle  string        `opa:"field:extra_single"`
		OPAMap pqx.JsonbMap  `opa:"field:extra_map"`
		Filter FilteredModel `gorm:"-" opa:"type:res"`
	}
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		model := &Model{
			TenantID:   "tenantId",
			TenantPath: pqx.UUIDArray{uuid.New(), uuid.New()},
			OwnerID:    uuid.New(),
			Sharing:    pqx.JsonbMap{"userid": []string{"read", "write"}},
			OPAArray:   []string{"v1", "v2"},
			OPASingle:  "value",
			OPAMap:     pqx.JsonbMap{"key": "value"},
			Filter:     FilteredModel{},
		}
		typ, v, e := ResolveResource(model)
		g.Expect(e).To(Succeed(), "resolve resource should not return error")
		assertResource(g, typ, v, &ExpectedResource{
			Type: "res",
			Values: map[string]interface{}{
				"tenant_id":    model.TenantID,
				"tenant_path":  toJsonSlice(model.TenantPath),
				"owner_id":     model.OwnerID.String(),
				"sharing":      toJsonObject(model.Sharing),
				"extra_array":  toJsonSlice(model.OPAArray),
				"extra_single": model.OPASingle,
				"extra_map":    toJsonObject(model.OPAMap),
			},
		})
	}
}

func SubTestResolveInvalidModel() test.GomegaSubTestFunc {
	type Model struct {
		ID         uuid.UUID     `gorm:"primaryKey;type:uuid;default:gen_random_uuid();"`
		OPAArray   []string      `gorm:"type:text[]" opa:"field:extra_array"`
		OPASingle  string        `opa:"field:extra_single"`
		OPAMap pqx.JsonbMap  `opa:"field:extra_map"`
		Filter FilteredModel `gorm:"-"`
	}
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		model := &Model{
			OPAArray:  []string{"v1", "v2"},
			OPASingle: "value",
			OPAMap:    pqx.JsonbMap{"key": "value"},
			Filter:    FilteredModel{},
		}
		_, _, e := ResolveResource(model)
		g.Expect(e).To(HaveOccurred(), "resolve resource should return error")
	}
}

func SubTestInvalidGenerics() test.GomegaSubTestFunc {
	type Model struct {
		ID         uuid.UUID     `gorm:"primaryKey;type:uuid;default:gen_random_uuid();"`
		OPAArray   []string      `gorm:"type:text[]" opa:"field:extra_array"`
		OPASingle  string        `opa:"field:extra_single"`
		OPAMap pqx.JsonbMap  `opa:"field:extra_map"`
		Filter FilteredModel `gorm:"-" opa:"res"`
	}
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		model := &Model{
			OPAArray:  []string{"v1", "v2"},
			OPASingle: "value",
			OPAMap:    pqx.JsonbMap{"key": "value"},
			Filter:    FilteredModel{},
		}
		models := []*Model{model}
		_, _, e := ResolveResource(&models)
		g.Expect(e).To(HaveOccurred(), "resolve resource should return error")
	}
}

/*************************
	Helper
 *************************/

type ExpectedResource struct {
	Type   string
	Values map[string]interface{}
}

func assertResource(g *gomega.WithT, resType string, resValues *opa.ResourceValues, expected *ExpectedResource) {
	g.Expect(resType).To(Equal(expected.Type), "resolved resource type should be correct")
	if len(expected.Values) == 0 {
		g.Expect(resValues).To(BeNil(), "resolved resource values should be nil")
		return
	}
	data, e := resValues.MarshalJSON()
	g.Expect(e).To(Succeed(), "marshalling resource values should not return error")
	var parsed map[string]interface{}
	g.Expect(json.Unmarshal(data, &parsed)).To(Succeed(), "json of resource values should be an object")
	g.Expect(parsed).To(HaveLen(len(expected.Values)), "resource values should have correct number of fields")
	for k, v := range expected.Values {
		g.Expect(parsed).To(HaveKey(k), "resource values should have key '%s'", k)
		g.Expect(parsed[k]).To(BeEquivalentTo(v), "resource values should have correct value of key '%s'", k)
	}
}

func toJsonSlice[T any](slice []T) (ret []interface{}) {
	ret = make([]interface{}, len(slice))
	for i := range slice {
		ret[i] = fmt.Sprint(slice[i])
	}
	return
}

func toJsonObject(obj interface{}) (ret interface{}) {
	jsonStr, e := json.Marshal(obj)
	if e != nil {
		return
	}
	_ = json.Unmarshal(jsonStr, &ret)
	return
}
