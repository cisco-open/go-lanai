package types

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/data/types/pqx"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/tenancy"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

type Audit struct {
	CreatedAt time.Time      `json:"createdAt,omitempty"`
	UpdatedAt time.Time      `json:"updatedAt,omitempty"`
	CreatedBy uuid.UUID      `type:"UUID;" json:"createdBy,omitempty"`
	UpdatedBy uuid.UUID      `type:"UUID;" json:"updatedBy,omitempty"`
}

type SoftDelete struct {
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleteAt,omitempty"`
}

type Tenancy struct {
	TenantID   uuid.UUID     `gorm:"type:UUID;not null"`
	TenantPath pqx.UUIDArray `gorm:"type:uuid[];index:,type:gin;not null"  json:"-"`
}

func (t *Tenancy) BeforeSave(tx *gorm.DB) error {
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
