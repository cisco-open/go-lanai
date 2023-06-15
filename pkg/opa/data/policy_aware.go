package opadata

import (
	"fmt"
	"gorm.io/gorm"
)

// PolicyAware is an embedded type for data model. It's responsible for applying PolicyFilter and
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

func (t PolicyAware) BeforeCreate(tx *gorm.DB) error {
	meta, e := loadMetadata(tx.Statement.Schema)
	if e != nil {
		return e
	}
	fmt.Println(meta)
	//if tenantId is not available
	//if t.TenantID == uuid.Nil {
	//	return errors.New("tenantId is required")
	//}
	//
	//if !shouldSkip(tx.Statement.Context, FilteringFlagWriteValueCheck, filteringModeDefault) && !security.HasAccessToTenant(tx.Statement.Context, t.TenantID.String()) {
	//	return errors.New(fmt.Sprintf("user does not have access to tenant %s", t.TenantID.String()))
	//}

	// TODO
	//path, err := tenancy.GetTenancyPath(tx.Statement.Context, t.TenantID.String())
	//if err == nil {
	//	t.TenantPath = path
	//}
	//return err
	return nil
}

// BeforeUpdate Check if OPA policy allow to update this model.
// We don't check the original values because we don't have that information in this hook. That check has to be done
// in application code.
func (t PolicyAware) BeforeUpdate(tx *gorm.DB) error {
	//dest := tx.Statement.Dest
	//tenantId, e := t.extractTenantId(tx.Statement.Context, dest)
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
	//	err = t.updateTenantPath(tx.Statement.Context, dest, path)
	//}
	//return err
	return nil
}