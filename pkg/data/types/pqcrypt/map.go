package pqcrypt

import (
	"context"
	"database/sql/driver"
	"github.com/google/uuid"
)

type EncryptedMap struct {
	EncryptedRaw
	Data  map[string]interface{} `json:"-"`
}

func NewEncryptedMap(kid uuid.UUID, v map[string]interface{}) *EncryptedMap {
	if kid == zeroUUID {
		return newEncryptedMap("", v)
	}
	return newEncryptedMap(kid.String(), v)
}

func newEncryptedMap(kid string, v map[string]interface{}) *EncryptedMap {
	return &EncryptedMap{
		EncryptedRaw: EncryptedRaw{
			KeyID: kid,
		},
		Data:  v,
	}
}

// Value implements driver.Valuer
func (d *EncryptedMap) Value() (driver.Value, error) {
	raw, e := Encrypt(context.Background(), d.KeyID, d.Data)
	if e != nil {
		return nil, e
	}
	d.EncryptedRaw = *raw
	return d.EncryptedRaw.Value()
}

// Scan implements sql.Scanner
func (d *EncryptedMap) Scan(src interface{}) error {
	if e := d.EncryptedRaw.Scan(src); e != nil {
		return e
	}

	return Decrypt(context.Background(), &d.EncryptedRaw, &d.Data)
}
