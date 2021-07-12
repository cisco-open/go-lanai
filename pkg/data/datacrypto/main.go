package datacrypto

import (
	"context"
)

var encryptor Encryptor = compositeEncryptor{plainTextEncryptor{}}

// Encrypt is a package level API that wraps shared Encryptor.Encrypt
func Encrypt(ctx context.Context, v interface{}, raw *EncryptedRaw) error {
	if encryptor == nil {
		return newEncryptionError("data encryption is not properly configured")
	}
	return encryptor.Encrypt(ctx, v, raw)
}

// Decrypt is a package level API that wraps shared Encryptor.Decrypt
func Decrypt(ctx context.Context, raw *EncryptedRaw, dest interface{}) error {
	if encryptor == nil {
		return newDecryptionError("data encryption is not properly configured")
	}
	return encryptor.Decrypt(ctx, raw, dest)
}
