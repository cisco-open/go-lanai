package types

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/data/types/pqx"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

type Audit struct {
	CreatedAt time.Time      `json:"createdAt,omitempty"`
	UpdatedAt time.Time      `json:"updatedAt,omitempty"`
	DeletedAt gorm.DeletedAt `json:"deleteAt,omitempty"`
	CreatedBy uuid.UUID      `type:"UUID;"  json:"createdBy,omitempty"`
	UpdatedBy uuid.UUID      `type:"UUID;" json:"updatedBy,omitempty"`
}

type Tenancy struct {
	TenantID   uuid.UUID     `gorm:"type:UUID;not null"`
	TenantPath pqx.UUIDArray `gorm:"type:uuid[];index:,type:gin;not null"  json:"-"`
}
