package filter_tests

import (
    "context"
    "cto-github.cisco.com/NFV-BU/go-lanai/pkg/data"
    "cto-github.cisco.com/NFV-BU/go-lanai/pkg/opa"
    opadata "cto-github.cisco.com/NFV-BU/go-lanai/pkg/opa/data"
    "errors"
    "github.com/google/uuid"
    "github.com/onsi/gomega"
    . "github.com/onsi/gomega"
    "gorm.io/gorm"
    "reflect"
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
	Asserts
 *************************/

func assertPostOpModel[T any](ctx context.Context, g *gomega.WithT, db *gorm.DB, id interface{}, op string, expectedKVs ...interface{}) *T {
    model, e := loadModel[T](ctx, db, id)
    if len(expectedKVs) == 0 {
        g.Expect(e).To(HaveOccurred(), "model should not exist after %s", op)
        g.Expect(errors.Is(e, data.ErrorRecordNotFound)).To(BeTrue(), "get model after %s should return record not found error", op)
        return nil
    }

    g.Expect(e).To(Succeed(), "model should exist after %s", op)
    rv := reflect.Indirect(reflect.ValueOf(model))
    for i := 0; i < len(expectedKVs)-1; i += 2 {
        k := expectedKVs[i].(string)
        fv := rv.FieldByName(k)
        g.Expect(fv.IsValid()).To(BeTrue(), `model should have field "%s"" (after %s)`, k, op)
        g.Expect(fv.Interface()).To(BeEquivalentTo(expectedKVs[i+1]), `model's field "%s" should have correct value (after %s)`, k, op)
    }
    return model
}

func assertDBResult(_ context.Context, g *gomega.WithT, rs *gorm.DB, op string, expectedErr error, expectedRows int) {
    defer func() {
        g.Expect(rs.RowsAffected).To(BeNumerically("==", expectedRows), "%s should affect %d rows", op, expectedRows)
    }()
    // if expected rows is 0, but actual result is opa.ErrAccessDenied, we consider it as acceptable behavior
    if expectedErr != nil {
        g.Expect(rs.Error).To(HaveOccurred(), "%s should return error", op)
        g.Expect(errors.Is(rs.Error, expectedErr)).To(BeTrue(), "%s should return correct error", op)
        return
    } else if expectedRows == 0 && rs.Error != nil {
        g.Expect(errors.Is(rs.Error, opa.ErrAccessDenied)).To(BeTrue(), "%s should return correct error", op)
        return
    } else {
        g.Expect(rs.Error).To(Succeed(), "%s should return no error", op)
    }
}

func assertOwnership[T any](g *gomega.WithT, ownerId uuid.UUID, op string, models ...*T) {
    for i, model := range models {
        rv := reflect.Indirect(reflect.ValueOf(model))
        fv := rv.FieldByName("OwnerID")
        g.Expect(fv.IsValid()).To(BeTrue(), `model should have field "OwnerID"`)
        g.Expect(fv.Interface()).To(BeEquivalentTo(ownerId), `model's' OwnerID should have correct value at idx %d' (%s)`, i, op)
    }
}

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
func loadModel[T any](ctx context.Context, db *gorm.DB, id interface{}, opts ...func(*gorm.DB)*gorm.DB) (*T, error) {
    var dest T
    tx := db.WithContext(ctx).Scopes(opadata.SkipFiltering())
    for _, fn := range opts {
        tx = fn(tx)
    }
    rs := tx.Take(&dest, id)
    if rs.Error != nil {
        return nil, rs.Error
    }
    return &dest, nil
}

// mustLoadModel load model without policy filtering
func mustLoadModel[T any](ctx context.Context, g *gomega.WithT, db *gorm.DB, id interface{}, opts ...func(*gorm.DB)*gorm.DB) *T {
    ret, e := loadModel[T](ctx, db, id, opts...)
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
