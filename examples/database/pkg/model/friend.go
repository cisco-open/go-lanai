package model

import (
	"github.com/google/uuid"
	"time"
)

type Friend struct {
	ID        uuid.UUID `gorm:"column:id;primary_key;type:UUID;default:gen_random_uuid();"`
	FirstName string    `gorm:"column:first_name;uniqueIndex:idx_friends_name;type:text;not null;"`
	LastName  string    `gorm:"column:last_name;uniqueIndex:idx_friends_name;type:text;not null;"`
	CreatedAt time.Time `json:"createdAt,omitempty"`
	UpdatedAt time.Time `json:"updatedAt,omitempty"`
	CreatedBy string    `gorm:"type:KeyID;" json:"createdBy,omitempty"`
	UpdatedBy string    `gorm:"type:KeyID;" json:"updatedBy,omitempty"`
}
