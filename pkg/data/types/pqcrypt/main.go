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
	"github.com/google/uuid"
)

const (
	errTmplNotConfigured = `data encryption is not properly configured`
)

var encryptor Encryptor = plainTextEncryptor{}

var zeroUUID = uuid.UUID{}

// Encrypt is a package level API that wraps shared Encryptor.Encrypt
func Encrypt(ctx context.Context, kid string, v interface{}) (*EncryptedRaw, error) {
	if encryptor == nil {
		return nil, newEncryptionError(errTmplNotConfigured)
	}
	return encryptor.Encrypt(ctx, kid, v)
}

// Decrypt is a package level API that wraps shared Encryptor.Decrypt
func Decrypt(ctx context.Context, raw *EncryptedRaw, dest interface{}) error {
	if encryptor == nil {
		return newDecryptionError(errTmplNotConfigured)
	}
	return encryptor.Decrypt(ctx, raw, dest)
}

// CreateKey create keys with given key ID.
// Note: KeyOptions is for future support, it's currently ignored
func CreateKey(ctx context.Context, kid string, opts ...KeyOptions) error {
	if encryptor == nil {
		return newEncryptionError(errTmplNotConfigured)
	}
	return encryptor.KeyOperations().Create(ctx, kid, opts...)
}

// CreateKeyWithUUID create keys with given key ID.
// Note: KeyOptions is for future support, it's currently ignored
func CreateKeyWithUUID(ctx context.Context, kid uuid.UUID, opts ...KeyOptions) error {
	if kid == zeroUUID {
		return CreateKey(ctx, "", opts...)
	}
	return CreateKey(ctx, kid.String(), opts...)
}
