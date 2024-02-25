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
    "fmt"
    "github.com/cisco-open/go-lanai/pkg/bootstrap"
    "github.com/cisco-open/go-lanai/pkg/utils"
    "github.com/pkg/errors"
    "strings"
)

const (
	PropertiesPrefix = "data.encryption"
)

type DataEncryptionProperties struct {
	Enabled bool          `json:"enabled"`
	Key     KeyProperties `json:"key"`
}

type KeyProperties struct {
	Type                 string `json:"type"`
	Exportable           bool   `json:"exportable"`
	AllowPlaintextBackup bool   `json:"allow-plaintext-backup"`
}

// https://www.vaultproject.io/api/secret/transit#create-key
const (
	KeyTypeAES128   = "aes128-gcm96"
	KeyTypeAES256   = "aes256-gcm96"
	KeyTypeChaCha20 = "chacha20-poly1305"
	KeyTypeED25519  = "ed25519"
	KeyTypeECDSA256 = "ecdsa-p256"
	KeyTypeECDSA384 = "ecdsa-p384"
	KeyTypeECDSA521 = "ecdsa-p521"
	KeyTypeRSA2048  = "rsa-2048"
	KeyTypeRSA3072  = "rsa-3072"
	KeyTypeRSA4096  = "rsa-4096"

	defaultKeyType = KeyTypeAES256
)

var supportedKeyTypes = utils.NewStringSet(
	KeyTypeAES128, KeyTypeAES256, KeyTypeChaCha20,
	KeyTypeED25519, KeyTypeECDSA256, KeyTypeECDSA384, KeyTypeECDSA521,
	KeyTypeRSA2048, KeyTypeRSA3072, KeyTypeRSA4096,
)

type KeyType string

// UnmarshalText implements encoding.TextUnmarshaler
func (t *KeyType) UnmarshalText(text []byte) error {
	str := strings.ToLower(strings.TrimSpace(string(text)))
	switch {
	case len(str) == 0:
		*t = defaultKeyType
	case supportedKeyTypes.Has(str):
		*t = KeyType(str)
	default:
		return fmt.Errorf("unknown encryption key type: %s", str)
	}
	return nil
}

//NewDataEncryptionProperties create a CockroachProperties with default values
func NewDataEncryptionProperties() *DataEncryptionProperties {
	return &DataEncryptionProperties{
		Enabled: false,
		Key: KeyProperties{
			Type:                 defaultKeyType,
			Exportable:           false,
			AllowPlaintextBackup: false,
		},
	}
}

//BindDataEncryptionProperties create and bind SessionProperties, with a optional prefix
func BindDataEncryptionProperties(ctx *bootstrap.ApplicationContext) DataEncryptionProperties {
	props := NewDataEncryptionProperties()
	if err := ctx.Config().Bind(props, PropertiesPrefix); err != nil {
		panic(errors.Wrap(err, "failed to bind DataEncryptionProperties"))
	}
	return *props
}
