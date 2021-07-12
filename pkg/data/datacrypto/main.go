package datacrypto

import (
	"context"
	"fmt"
)

var encryptor Encryptor = plainTextEncryptor{}

// Encrypt is a package level API that wraps shared Encryptor.Encrypt
func Encrypt(ctx context.Context, v interface{}, raw *EncryptedRaw) error {
	if encryptor == nil {
		return fmt.Errorf("data encryption is not properly configured")
	}
	return encryptor.Encrypt(ctx, v, raw)
}

// Decrypt is a package level API that wraps shared Encryptor.Decrypt
func Decrypt(ctx context.Context, raw *EncryptedRaw, v interface{}) error {
	if encryptor == nil {
		return fmt.Errorf("data encryption is not properly configured")
	}
	return encryptor.Decrypt(ctx, raw, v)
}
