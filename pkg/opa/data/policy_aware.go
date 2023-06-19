package opadata

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/opa"
	"gorm.io/gorm"
)

// PolicyAware is an embedded type for data targetModel. It's responsible for applying PolicyFilter and
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
	meta, e := loadMetadata(tx.Statement.Schema)
	if e != nil {
		// TODO proper error
		return e
	}

	if shouldSkip(tx.Statement.Context, PolicyFlagCreate, meta.Mode) {
		return nil
	}

	// TODO TBD: should we auto-populate tenant ID, tenant path, owner, etc

	// Note: enforce policy is performed in PolicyFilter
	return nil
}

// BeforeUpdate Check if OPA policy allow to update this policy related field.
// We don't check the original values because we don't have that information in this hook.
func (p PolicyAware) BeforeUpdate(tx *gorm.DB) error {
	// TODO TBD: should we check tenant ID, tenant path, owner, etc ?
	// 			 should we auto-populate tenant ID, tenant path, owner, etc ?
	return nil
}

/*******************
	Helpers
 *******************/

func checkPolicy(ctx context.Context, m *targetModel, op opa.ResourceOperation) error {
	input := map[string]interface{}{}
	for k, tagged := range m.meta.Fields {
		v := m.val.FieldByIndex(tagged.StructField.Index).Interface()
		input[k] = v
	}
	return opa.AllowResource(ctx, m.meta.ResType, op, func(res *opa.Resource) {
		res.ExtraData = input
	})
}

