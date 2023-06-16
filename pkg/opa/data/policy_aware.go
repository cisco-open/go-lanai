package opadata

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/opa"
	"fmt"
	"gorm.io/gorm"
	"reflect"
)

// PolicyAware is an embedded type for data modelInfo. It's responsible for applying PolicyFilter and
// populating/checking OPA policy related data field
// TODO update following description
// when crating/updating. PolicyAware implements
// - callbacks.BeforeCreateInterface
// - callbacks.BeforeUpdateInterface
// When used as an embedded type, tag `filter` can be used to override default tenancy check behavior:
// - `filter:"w"`: 	create/update/delete are enforced (Default mode)
// - `filter:"rw"`: CRUD operations are all enforced,
//					this mode filters result of any Select/Update/Delete query based on current security context
// - `filter:"-"`: 	filtering is disabled. Note: setting TenantID to in-accessible tenant is still enforced.
//					to disable TenantID value check, use SkipPolicyFiltering
// e.g.
// <code>
// type TenancyModel struct {
//		ID         uuid.UUID `gorm:"primaryKey;type:uuid;default:gen_random_uuid();"`
//		Tenancy    `filter:"rw"`
// }
// </code>
type PolicyAware struct {
	OPAPolicyFilter PolicyFilter `gorm:"-"`
}

func (p PolicyAware) BeforeCreate(tx *gorm.DB) error {
	m, e := p.collectModelInfo(tx)
	if e != nil {
		return e
	}
	// enforce policy
	if e := p.checkPolicy(tx.Statement.Context, &m, opa.OpCreate); e != nil {
		return e
	}


	//if tenantId is not available
	//if p.TenantID == uuid.Nil {
	//	return errors.New("tenantId is required")
	//}
	//
	//if !shouldSkip(tx.Statement.Context, FilteringFlagWriteValueCheck, filteringModeDefault) && !security.HasAccessToTenant(tx.Statement.Context, p.TenantID.String()) {
	//	return errors.New(fmt.Sprintf("user does not have access to tenant %s", p.TenantID.String()))
	//}

	// TODO
	//path, err := tenancy.GetTenancyPath(tx.Statement.Context, p.TenantID.String())
	//if err == nil {
	//	p.TenantPath = path
	//}
	//return err
	return nil
}

// BeforeUpdate Check if OPA policy allow to update this modelInfo.
// We don't check the original values because we don't have that information in this hook. That check has to be done
// in application code.
func (p PolicyAware) BeforeUpdate(tx *gorm.DB) error {
	_, e := p.collectModelInfo(tx)
	if e != nil {
		return e
	}
	//dest := tx.Statement.Dest
	//tenantId, e := p.extractTenantId(tx.Statement.Context, dest)
	//if e != nil || tenantId == uuid.Nil {
	//	return e
	//}
	//
	//if !shouldSkip(tx.Statement.Context, FilteringFlagWriteValueCheck, filteringModeDefault) && !security.HasAccessToTenant(tx.Statement.Context, tenantId.String()) {
	//	return errors.New(fmt.Sprintf("user does not have access to tenant %s", tenantId.String()))
	//}

	// TODO
	//path, err := tenancy.GetTenancyPath(tx.Statement.Context, tenantId.String())
	//if err == nil {
	//	err = p.updateTenantPath(tx.Statement.Context, dest, path)
	//}
	//return err
	return nil
}

/*******************
	Helpers
 *******************/

// modelInfo collected information about current modelInfo
type modelInfo struct {
	meta *metadata
	ptr  reflect.Value
	val  reflect.Value
}

func (p PolicyAware) checkPolicy(ctx context.Context, m *modelInfo, op opa.ResourceOperation) error {
	input := map[string]interface{}{}
	for k, tagged := range m.meta.Fields {
		v := m.val.FieldByIndex(tagged.StructField.Index).Interface()
		input[k] = v
	}
	return opa.AllowResource(ctx, m.meta.ResType, op, func(res *opa.Resource) {
		res.ExtraData = input
	})
}

func (p PolicyAware) collectModelInfo(tx *gorm.DB) (m modelInfo, err error) {
	m.meta, err = loadMetadata(tx.Statement.Schema)
	if err != nil {
		return
	}
	m.ptr, err = p.resolveModel(tx)
	if err != nil {
		return
	}
	m.val = m.ptr.Elem()

	// sanity check
	if m.meta.Schema.ModelType != m.val.Type() {
		return modelInfo{}, fmt.Errorf("policy metadata is not found")
	}
	return

}

// resolveModel find the pointer of enclosing modelInfo struct
func (p PolicyAware) resolveModel(tx *gorm.DB) (rv reflect.Value, err error) {
	switch tx.Statement.ReflectValue.Kind() {
	case reflect.Slice, reflect.Array:
		if tx.Statement.CurDestIndex >= tx.Statement.ReflectValue.Len() {
			break
		}
		rv = tx.Statement.ReflectValue.Index(tx.Statement.CurDestIndex)
	case reflect.Struct:
		rv = tx.Statement.ReflectValue
	}
	switch {
	case !rv.IsValid():
		break
	case rv.Type() == tx.Statement.Schema.ModelType && rv.CanAddr():
		return rv.Addr(), nil
	case rv.Type() == reflect.PointerTo(tx.Statement.Schema.ModelType):
		return rv, nil
	}
	return rv, fmt.Errorf("unable to extract modelInfo value")
}
