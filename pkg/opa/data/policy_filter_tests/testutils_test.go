package policy_filter_tests

import (
    "context"
    opadata "cto-github.cisco.com/NFV-BU/go-lanai/pkg/opa/data"
    "github.com/google/uuid"
    "github.com/onsi/gomega"
    . "github.com/onsi/gomega"
    "gorm.io/gorm"
)

type LookupKey struct {
    Tenant uuid.UUID
    Owner  uuid.UUID
}

var (
    MockedModelLookupByTenant = map[uuid.UUID][]uuid.UUID{}
    MockedModelLookupByOwner  = map[uuid.UUID][]uuid.UUID{}
    MockedModelLookupByKey    = map[LookupKey][]uuid.UUID{}
)

/*************************
	Helpers
 *************************/

func resetIdLookup() {
    MockedModelLookupByTenant = map[uuid.UUID][]uuid.UUID{}
    MockedModelLookupByOwner = map[uuid.UUID][]uuid.UUID{}
    MockedModelLookupByKey = map[LookupKey][]uuid.UUID{}
}

func prepareIdLookup(modelId uuid.UUID, key LookupKey) {
    var ids []uuid.UUID
    ids, _ = MockedModelLookupByKey[key]
    MockedModelLookupByKey[key] = appendOrNew(ids, modelId)
    ids, _ = MockedModelLookupByTenant[key.Tenant]
    MockedModelLookupByTenant[key.Tenant] = appendOrNew(ids, modelId)
    ids, _ = MockedModelLookupByOwner[key.Owner]
    MockedModelLookupByOwner[key.Owner] = appendOrNew(ids, modelId)
}

func findID(ownerId, tenantId uuid.UUID) uuid.UUID {
    key := LookupKey{Tenant: tenantId, Owner: ownerId}
    ids, _ := MockedModelLookupByKey[key]
    for i := range ids {
        if ids[i] == uuid.Nil {
            continue
        }
        ret := ids[i]
        ids[i] = uuid.Nil
        return ret
    }
    return uuid.Nil
}

func findIDByTenant(tenantId uuid.UUID) uuid.UUID {
    ids, _ := MockedModelLookupByTenant[tenantId]
    for i := range ids {
        if ids[i] == uuid.Nil {
            continue
        }
        ret := ids[i]
        ids[i] = uuid.Nil
        return ret
    }
    return uuid.Nil
}

func findIDByOwner(ownerId uuid.UUID) uuid.UUID {
    ids, _ := MockedModelLookupByOwner[ownerId]
    for i := range ids {
        if ids[i] == uuid.Nil {
            continue
        }
        ret := ids[i]
        ids[i] = uuid.Nil
        return ret
    }
    return uuid.Nil
}

// loadModel load model without policy filtering
func loadModel[T any](ctx context.Context, db *gorm.DB, id uuid.UUID) (*T, error) {
    var dest T
    rs := db.WithContext(ctx).Scopes(opadata.SkipPolicyFiltering()).Take(&dest, id)
    if rs.Error != nil {
        return nil, rs.Error
    }
    return &dest, nil
}

// mustLoadModel load model without policy filtering
func mustLoadModel[T any](ctx context.Context, g *gomega.WithT, db *gorm.DB, id uuid.UUID) *T {
    ret, e := loadModel[T](ctx, db, id)
    g.Expect(e).To(Succeed(), "model must exists")
    return ret
}

func appendOrNew[T any](slice []T, values ...T) []T {
    if slice == nil {
        slice = make([]T, 0, 5)
    }
    slice = append(slice, values...)
    return slice
}

func shallowCopyMap[K comparable, V any](src map[K]V) map[K]V {
    dest := map[K]V{}
    for k, v := range src {
        dest[k] = v
    }
    return dest
}
