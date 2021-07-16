package pqcrypt

import (
	"context"
)

type plainTextEncryptor struct{}

func (enc plainTextEncryptor) Encrypt(ctx context.Context, kid string, v interface{}) (raw *EncryptedRaw, err error) {
	raw = &EncryptedRaw{
		Ver:   V2,
		KeyID: kid,
		Alg:   AlgPlain,
	}
	switch {
	case raw.KeyID == "":
		return nil, newEncryptionError("KeyID is required for algorithm %v", raw.Alg)
	}

	raw.Raw = v
	return
}

func (enc plainTextEncryptor) Decrypt(_ context.Context, raw *EncryptedRaw, dest interface{}) error {
	if raw == nil {
		return newDecryptionError("raw data is nil")
	}

	switch raw.Ver {
	case V1, V2:
		if raw.Alg != AlgPlain {
			return ErrUnsupportedAlgorithm
		}
		if e := tryAssign(raw.Raw, dest); e != nil {
			return e
		}
	default:
		return ErrUnsupportedVersion
	}
	return nil
}

func (enc plainTextEncryptor) KeyOperations() KeyOperations {
	return noopKeyOps
}


