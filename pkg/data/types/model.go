package types

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
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
