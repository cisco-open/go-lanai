package opadata

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/tenancy"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/reflectutils"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"reflect"
)

const (
	fieldTenantID   = "TenantID"
	fieldTenantPath = "PolicyFilter"
	colTenantID     = "tenant_id"
	colTenantPath   = "tenant_path"
)

var (
	typeUUID          = reflect.TypeOf(uuid.Nil)
	typeTenantPath    = reflect.TypeOf(PolicyFilter{})
	typeTenancy    = reflect.TypeOf(Tenancy{})
	typeTenancyPtr    = reflect.TypeOf(&Tenancy{})
	mapKeysTenantID   = utils.NewStringSet(fieldTenantID, colTenantID)
	mapKeysTenantPath = utils.NewStringSet(fieldTenantPath, colTenantPath)
)

type ckTenancyCheckMode struct{}

const (
	TenancyCheckFlagWriteValueCheck TenancyCheckFlag = 1 << iota
	TenancyCheckFlagWriteFiltering
	TenancyCheckFlagReadFiltering
)

// TenancyCheckFlag bitwise Flag of tenancy flag mode
type TenancyCheckFlag uint

const (
	tcModeDefault = tcMode(TenancyCheckFlagWriteFiltering | TenancyCheckFlagWriteValueCheck)
)

// tcMode enum of tenancyCheckMode
type tcMode uint

func (m tcMode) hasFlags(flags ...TenancyCheckFlag) bool {
	for _, flag := range flags {
		if m & tcMode(flag) == 0 {
			return false
		}
	}
	return true
}

// SkipTenancyCheck is used as a scope for gorm.DB to skip tenancy check
// e.g. db.WithContext(ctx).Scopes(SkipTenancyCheck()).Find(...)
// Note using this scope without context would panic
func SkipTenancyCheck() func(*gorm.DB) *gorm.DB {
	return TenancyCheck(0)
}

// TenancyCheck is used as a scope for gorm.DB to override tenancy check
// e.g. db.WithContext(ctx).Scopes(TenancyCheck()).Find(...)
// Note using this scope without context would panic
func TenancyCheck(flags ...TenancyCheckFlag) func(*gorm.DB) *gorm.DB {
	return func(tx *gorm.DB) *gorm.DB {
		if tx.Statement.Context == nil {
			panic("SkipTenancyCheck used without context")
		}
		var mode tcMode
		for _, flag := range flags {
			mode = mode | tcMode(flag)
		}
		ctx := context.WithValue(tx.Statement.Context, ckTenancyCheckMode{}, mode)
		tx.Statement.Context = ctx
		return tx
	}
}

// Tenancy is an embedded type for data model. It's responsible for populating PolicyFilter and check for Tenancy related data
// when crating/updating. Tenancy implements
// - callbacks.BeforeCreateInterface
// - callbacks.BeforeUpdateInterface
// When used as an embedded type, tag `filter` can be used to override default tenancy check behavior:
// - `filter:"w"`: 	create/update/delete are enforced (Default mode)
// - `filter:"rw"`: CRUD operations are all enforced,
//					this mode filters result of any Select/Update/Delete query based on current security context
// - `filter:"-"`: 	filtering is disabled. Note: setting TenantID to in-accessible tenant is still enforced.
//					to disable TenantID value check, use SkipTenancyCheck
// e.g.
// <code>
// type TenancyModel struct {
//		ID         uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid();"`
//		Tenancy    `filter:"rw"`
// }
// </code>
type Tenancy struct {
	TenantID   uuid.UUID    `gorm:"type:KeyID;not null"`
	TenantPath PolicyFilter `gorm:"type:uuid[];index:,type:gin;not null"  json:"-"`
}

// SkipTenancyCheck is used for embedding models to override tenancy check behavior.
// It should be called within model's hooks. this function would panic if context is not set yet
func (Tenancy) SkipTenancyCheck(tx *gorm.DB) {
	SkipTenancyCheck()(tx)
}

func (t *Tenancy) BeforeCreate(tx *gorm.DB) error {
	//if tenantId is not available
	if t.TenantID == uuid.Nil {
		return errors.New("tenantId is required")
	}

	if !shouldSkip(tx.Statement.Context, TenancyCheckFlagWriteValueCheck, tcModeDefault) && !security.HasAccessToTenant(tx.Statement.Context, t.TenantID.String()) {
		return errors.New(fmt.Sprintf("user does not have access to tenant %s", t.TenantID.String()))
	}

	path, err := tenancy.GetTenancyPath(tx.Statement.Context, t.TenantID.String())
	if err == nil {
		t.TenantPath = path
	}
	return err
}

// BeforeUpdate Check if user is allowed to update this item's tenancy to the target tenant.
// (i.e. if user has access to the target tenant)
// We don't check the original tenancy because we don't have that information in this hook. That check has to be done
// in application code.
func (t *Tenancy) BeforeUpdate(tx *gorm.DB) error {
	dest := tx.Statement.Dest
	tenantId, e := t.extractTenantId(tx.Statement.Context, dest)
	if e != nil || tenantId == uuid.Nil {
		return e
	}

	if !shouldSkip(tx.Statement.Context, TenancyCheckFlagWriteValueCheck, tcModeDefault) && !security.HasAccessToTenant(tx.Statement.Context, tenantId.String()) {
		return errors.New(fmt.Sprintf("user does not have access to tenant %s", tenantId.String()))
	}

	path, err := tenancy.GetTenancyPath(tx.Statement.Context, tenantId.String())
	if err == nil {
		err = t.updateTenantPath(tx.Statement.Context, dest, path)
	}
	return err
}

func (t Tenancy) extractTenantId(_ context.Context, dest interface{}) (uuid.UUID, error) {
	v := reflect.ValueOf(dest)
	for ; v.Kind() == reflect.Ptr; v = v.Elem() {
		// SuppressWarnings go:S108 empty block is intended
	}

	switch v.Kind() {
	case reflect.Map:
		if v.Type().Key().Kind() != reflect.String {
			return uuid.Nil, fmt.Errorf("unsupported gorm update target type [%T], please use struct ptr, struct or map", dest)
		}
		if _, ev, ok := t.findMapValue(v, mapKeysTenantID, typeUUID); ok {
			return ev.Interface().(uuid.UUID), nil
		}
	case reflect.Struct:
		_, fv, ok := t.findStructField(v, fieldTenantID, typeUUID)
		if ok {
			return fv.Interface().(uuid.UUID), nil
		}
	default:
		return uuid.Nil, fmt.Errorf("unsupported gorm update target type [%T], please use struct ptr, struct or map", dest)
	}
	return uuid.Nil, nil
}

func (t *Tenancy) updateTenantPath(_ context.Context, dest interface{}, tenancyPath PolicyFilter) error {
	v := reflect.ValueOf(dest)
	if v.Kind() == reflect.Struct {
		return fmt.Errorf("cannot update tenancy automatically to %T, please use struct ptr or map", dest)
	}
	for ; v.Kind() == reflect.Ptr; v = v.Elem() {
		// SuppressWarnings go:S108 empty block is intended
	}

	switch v.Kind() {
	case reflect.Map:
		if v.Type().Key().Kind() != reflect.String {
			return fmt.Errorf("cannot update tenancy automatically with gorm update target type [%T], please use struct ptr or map", dest)
		}
		ek, ev, ok := t.findMapValue(v, mapKeysTenantPath, typeTenantPath)
		// Note: if tenant path is explicitly set and correct, we don't change it
		switch {
		case ok && !reflect.DeepEqual(ev.Interface(), tenancyPath):
			return fmt.Errorf("incorrect %s was set to gorm update target map", ek)
		case !ok:
			v.SetMapIndex(reflect.ValueOf(fieldTenantPath), reflect.ValueOf(tenancyPath))
		}
	case reflect.Struct:
		if _, fv, ok := t.findStructField(v, fieldTenantPath, typeTenantPath); ok {
			fv.Set(reflect.ValueOf(tenancyPath))
		}
	default:
		return errors.New("cannot update tenancy automatically, please use struct ptr or map as gorm update target value")
	}
	return nil
}

func (Tenancy) findStructField(sv reflect.Value, name string, ft reflect.Type) (f reflect.StructField, fv reflect.Value, ok bool) {
	f, ok = reflectutils.FindStructField(sv.Type(), func(t reflect.StructField) bool {
		return t.Name == name && ft.AssignableTo(t.Type)
	})
	if ok {
		fv = sv.FieldByIndex(f.Index)
	}
	return
}

func (Tenancy) findMapValue(mv reflect.Value, keys utils.StringSet, ft reflect.Type) (string, reflect.Value, bool) {
	for iter := mv.MapRange(); iter.Next(); {
		k := iter.Key().String()
		if !keys.Has(k) {
			continue
		}
		v := iter.Value()
		if !v.IsZero() && ft.AssignableTo(v.Type()) {
			return k, v, true
		}
	}
	return "", reflect.Value{}, false
}

func shouldSkip(ctx context.Context, flag TenancyCheckFlag, fallback tcMode) bool {
	if ctx == nil || !security.IsFullyAuthenticated(security.Get(ctx)) {
		return true
	}
	switch v := ctx.Value(ckTenancyCheckMode{}).(type) {
	case tcMode:
		return !v.hasFlags(flag)
	default:
		return !fallback.hasFlags(flag)
	}
}
