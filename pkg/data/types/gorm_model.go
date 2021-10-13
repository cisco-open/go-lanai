package types

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/data/types/pqx"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/tenancy"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"reflect"
	"time"
)

type Audit struct {
	CreatedAt time.Time      `json:"createdAt,omitempty"`
	UpdatedAt time.Time      `json:"updatedAt,omitempty"`
	CreatedBy uuid.UUID      `type:"KeyID;" json:"createdBy,omitempty"`
	UpdatedBy uuid.UUID      `type:"KeyID;" json:"updatedBy,omitempty"`
}

type SoftDelete struct {
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleteAt,omitempty"`
}

var logger = log.New("tenancy")

type Tenancy struct {
	TenantID   uuid.UUID     `gorm:"type:KeyID;not null"`
	TenantPath pqx.UUIDArray `gorm:"type:uuid[];index:,type:gin;not null"  json:"-"`
}

func (t *Tenancy) BeforeCreate(tx *gorm.DB) error {
	//if tenantId is not available
	if t.TenantID == uuid.Nil {
		return errors.New("tenantId is required")
	}

	if !security.HasAccessToTenant(tx.Statement.Context, t.TenantID.String()) {
		return errors.New(fmt.Sprintf("user does not have access to tenant %s", t.TenantID.String()))
	}

	path, err := tenancy.GetTenancyPath(tx.Statement.Context, t.TenantID.String())

	if err != nil {
		return err
	}

	t.TenantPath = path

	return nil
}

// BeforeUpdate Check if user is allowed to update this item's tenancy to the target tenant.
// (i.e. if user has access to the target tenant)
// We don't check the original tenancy because we don't have that information in this hook. That check has to be done
// in application code.
func (t *Tenancy) BeforeUpdate(tx *gorm.DB) error {
	dest := tx.Statement.Dest
	tenantId, err := t.getTenantId(tx.Statement.Context, dest)

	if err != nil {
		return err
	}

	logger.Debugf("target tenancy is %v", tenantId)

	//not updating tenantId, bail
	if tenantId == uuid.Nil {
		return nil
	}

	if !security.HasAccessToTenant(tx.Statement.Context, tenantId.String()) {
		return errors.New(fmt.Sprintf("user does not have access to tenant %s", t.TenantID.String()))
	}

	path, err := tenancy.GetTenancyPath(tx.Statement.Context, tenantId.String())

	if err != nil {
		return err
	}

	err = t.updateTenantPath(tx.Statement.Context, dest, path)
	return err
}

func (t *Tenancy) BeforeDelete(tx *gorm.DB) error {
	//TODO: possible to add where clause to query?

	//TODO: if tenantId changed, check if user has access to target tenantId

	return nil
}

func (t *Tenancy) getTenantId(ctx context.Context, dest interface{}) (tenantId uuid.UUID, err error) {
	fieldName := "TenantID"

	v := reflect.ValueOf(dest)

	if v.Kind() == reflect.Ptr {
		elem := v.Elem()
		var fv reflect.Value
		if elem.Kind() == reflect.Ptr { //TODO: loop first
			logger.WithContext(ctx).Warnf("A pointer to a pointer was used as an update destination.")
			fv = elem.Elem().FieldByName(fieldName) //TODO: may need to recover if it's ptr of ptr of ptr
		} else {
			fv = elem.FieldByName(fieldName)
		}
		if !fv.IsZero() {
			tenantId, _ = fv.Interface().(uuid.UUID)
		}
	} else if v.Kind() == reflect.Struct {
		fv := v.FieldByName(fieldName)
		if !fv.IsZero() {
			tenantId, _ = fv.Interface().(uuid.UUID)
		}
	} else if v.Kind() == reflect.Map {
		for _, mkey := range v.MapKeys() {
			if mkey.Kind() == reflect.String && mkey.String() == fieldName {
				fv := v.MapIndex(mkey)
				if !fv.IsZero() {
					tenantId, _ = fv.Interface().(uuid.UUID)
				}
			}
		}
	} else {
		return tenantId,errors.New("unsupported gorm update target value, please use ptr, struct or map")
	}
	return tenantId, nil
}

func (t *Tenancy) updateTenantPath(ctx context.Context, dest interface{}, tenancyPath []uuid.UUID) error {
	fieldName := "TenantPath"

	v := reflect.ValueOf(dest)

	if v.Kind() == reflect.Ptr {
		elem := v.Elem()
		if elem.Kind() == reflect.Ptr {
			logger.WithContext(ctx).Warnf("A pointer to a pointer was used as an update destination.")
			elem = elem.Elem() //TODO: may need to recover if it's ptr of ptr of ptr
		}
		fv := elem.FieldByName(fieldName)
		fv.Set(reflect.ValueOf(tenancyPath))
	} else if v.Kind() == reflect.Map {
		for _, mkey := range v.MapKeys() {
			if mkey.Kind() == reflect.String && mkey.String() == fieldName {
				v.SetMapIndex(mkey, reflect.ValueOf(tenancyPath))
			}
		}
	} else {
		return errors.New("cannot update tenancy automatically, please use ptr or map as gorm update target value")
	}
	return nil
}
