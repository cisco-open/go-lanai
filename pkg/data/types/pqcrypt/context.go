// Copyright 2023 Cisco Systems, Inc. and its affiliates
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package pqcrypt

import (
    "context"
    "database/sql/driver"
    "encoding/json"
    "fmt"
    "github.com/cisco-open/go-lanai/pkg/data"
    "github.com/cisco-open/go-lanai/pkg/data/types/pqx"
    "strings"
)

var (
	ErrUnsupportedVersion   = data.NewDataError(data.ErrorCodeOrmMapping, "unsupported version of encrypted data format")
	ErrUnsupportedAlgorithm = data.NewDataError(data.ErrorCodeOrmMapping, "unsupported encryption algorithm of data")
	ErrInvalidFormat        = data.NewDataError(data.ErrorCodeOrmMapping, "invalid encrypted data")
	ErrInvalidV1Format        = data.NewDataError(data.ErrorCodeOrmMapping, "invalid V1 data payload format")
)

/*************************
	Enums
 *************************/

const (
	// V1 is Java compatible data structure
	V1 Version = 1
	// V2 is Generic JSON version, default format of go-lanai
	V2 Version = 2

	defaultVersion = V2
	minVersion     = V2
	v1Text         = "1"
	v2Text         = "2"
)

type Version int

// UnmarshalText implements encoding.TextUnmarshaler
func (v *Version) UnmarshalText(text []byte) error {
	str := strings.TrimSpace(string(text))
	switch str {
	case v1Text:
		*v = V1
	case v2Text:
		*v = V2
	case "":
		*v = defaultVersion
	default:
		return fmt.Errorf("unknown encrypted data version: %s", str)
	}
	return nil
}

// UnmarshalJSON implements json.Unmarshaler with V1 support
func (v *Version) UnmarshalJSON(data []byte) (err error) {
	var i int
	if e := json.Unmarshal(data, &i); e != nil {
		return e
	}

	switch i {
	case int(V1), int(V2):
		*v = Version(i)
	case 0:
		*v = defaultVersion
	default:
		return fmt.Errorf("unknown encrypted data version: %d", i)
	}
	return nil
}

const (
	AlgPlain   Algorithm = "p"
	AlgVault   Algorithm = "e" // this value is compatible with Java counterpart
	defaultAlg           = AlgPlain
)

type Algorithm string

// UnmarshalText implements encoding.TextUnmarshaler
func (a *Algorithm) UnmarshalText(text []byte) error {
	str := strings.TrimSpace(string(text))
	switch str {
	case string(AlgPlain):
		*a = AlgPlain
	case string(AlgVault):
		*a = AlgVault
	case "":
		*a = defaultAlg
	default:
		return fmt.Errorf("unknown encrypted data algorithm: %s", str)
	}
	return nil
}

/*************************
	Data
 *************************/

// EncryptedRaw is the carrier of encrypted data
// this data type implements gorm.Valuer, schema.GormDataTypeInterface
type EncryptedRaw struct {
	Ver   Version     `json:"v"`
	KeyID string      `json:"kid,omitempty"`
	Alg   Algorithm   `json:"alg,omitempty"`
	Raw   json.RawMessage `json:"d,omitempty"`
}

// GormDataType implements schema.GormDataTypeInterface
func (EncryptedRaw) GormDataType() string {
	return "jsonb"
}

// Value implements driver.Valuer
func (d *EncryptedRaw) Value() (driver.Value, error) {
	//we need to check nil here instead of in the JsonbValue method
	//because the input to JsonbValue is interface{}. Since d has a type
	// the v==nil check in JsonbValue won't return true
	if d == nil {
		return nil, nil
	}
	return pqx.JsonbValue(d)
}

// Scan implements sql.Scanner
func (d *EncryptedRaw) Scan(src interface{}) error {
	return pqx.JsonbScan(src, d)
}

/*************************
	Interface
 *************************/

type Encryptor interface {
	// Encrypt encrypt given "v" and populate EncryptedRaw.Raw
	// The process may read EncryptedRaw.Alg and EncryptedRaw.KeyID and update EncryptedRaw.Ver
	Encrypt(ctx context.Context, kid string, v interface{}) (*EncryptedRaw, error)

	// Decrypt reads EncryptedRaw and populate the decrypted data into given "v"
	// if v is not pointer type, this method may return error
	Decrypt(ctx context.Context, raw *EncryptedRaw, dest interface{}) error

	// KeyOperations returns an object that operates on keys.
	// depending on configurations, this method may returns no-op impl, but never nil
	KeyOperations() KeyOperations
}

type KeyOptions func(opt *keyOption)
type keyOption struct {
	ktype            string
	exportable       bool
	allowPlaintextBk bool
}

type KeyOperations interface {
	// Create create keys with given key ID.
	// Note: KeyOptions is for future support, it's currently ignored
	Create(ctx context.Context, kid string, opts ...KeyOptions) error
}

/*************************
	Common
 *************************/

type compositeEncryptor []Encryptor

// Encrypt always uses first Encryptor
func (enc compositeEncryptor) Encrypt(ctx context.Context, kid string, v interface{}) (*EncryptedRaw, error) {
	if len(enc) != 0 {
		return enc[0].Encrypt(ctx, kid, v)
	}
	return nil, newEncryptionError("encryptor is not properly configured")
}

func (enc compositeEncryptor) Decrypt(ctx context.Context, raw *EncryptedRaw, dest interface{}) error {
	for _, delegate := range enc {
		e := delegate.Decrypt(ctx, raw, dest)
		switch e { //nolint:errorlint // special error is global var
		case nil:
			return nil
		case ErrUnsupportedAlgorithm, ErrUnsupportedVersion:
			continue
		default:
			return e
		}
	}
	return newDecryptionError("encryptor is not available for ver=%d and alg=%v", raw.Ver, raw.Alg)
}

func (enc compositeEncryptor) KeyOperations() KeyOperations {
	ret := make(compositeKeyOperations, 0, len(enc))
	for _, delegate := range enc {
		ops := delegate.KeyOperations()
		if ops == noopKeyOps {
			continue
		}
		ret = append(ret, ops)
	}
	return ret
}

type compositeKeyOperations []KeyOperations

func (o compositeKeyOperations) Create(ctx context.Context, kid string, opts ...KeyOptions) error {
	for _, ops := range o {
		if e := ops.Create(ctx, kid, opts...); e != nil {
			return e
		}
	}
	return nil
}

type noopKeyOperations struct{}

var noopKeyOps = noopKeyOperations{}

func (o noopKeyOperations) Create(_ context.Context, _ string, _ ...KeyOptions) error {
	return nil
}
