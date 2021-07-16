package pqcrypt

import (
	"context"
	"github.com/google/uuid"
)

var encryptor Encryptor = plainTextEncryptor{}

var zeroUUID = uuid.UUID{}

// Encrypt is a package level API that wraps shared Encryptor.Encrypt
func Encrypt(ctx context.Context, kid string, v interface{}) (*EncryptedRaw, error) {
	if encryptor == nil {
		return nil, newEncryptionError("data encryption is not properly configured")
	}
	return encryptor.Encrypt(ctx, kid, v)
}

// Decrypt is a package level API that wraps shared Encryptor.Decrypt
func Decrypt(ctx context.Context, raw *EncryptedRaw, dest interface{}) error {
	if encryptor == nil {
		return newDecryptionError("data encryption is not properly configured")
	}
	return encryptor.Decrypt(ctx, raw, dest)
}

// CreateKey create keys with given key ID.
// Note: KeyOptions is for future support, it's currently ignored
func CreateKey(ctx context.Context, kid string, opts ...KeyOptions) error {
	if encryptor == nil {
		return newEncryptionError("data encryption is not properly configured")
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
