package pqcrypt

import (
	"context"
	"database/sql/driver"
	"github.com/google/uuid"
)

type EncryptedMap struct {
	EncryptedRaw
	Data map[string]interface{} `json:"-"`
}

func NewEncryptedMap(kid uuid.UUID, alg Algorithm, v map[string]interface{}) *EncryptedMap {
	return newEncryptedMap(V2, kid, alg, v)
}

func newEncryptedMap(ver Version, kid uuid.UUID, alg Algorithm, v map[string]interface{}) *EncryptedMap {
	return &EncryptedMap{
		EncryptedRaw: EncryptedRaw{
			Ver:   ver,
			KeyID: kid,
			Alg:   alg,
		},
		Data: v,
	}
}

// Value implements driver.Valuer
func (d *EncryptedMap) Value() (driver.Value, error) {
	if e := Encrypt(context.Background(), d.Data, &d.EncryptedRaw); e != nil {
		return nil, e
	}
	return d.EncryptedRaw.Value()
}

// Scan implements sql.Scanner
func (d *EncryptedMap) Scan(src interface{}) error {
	if e := d.EncryptedRaw.Scan(src); e != nil {
		return e
	}

	return Decrypt(context.Background(), &d.EncryptedRaw, &d.Data)
}
