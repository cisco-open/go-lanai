package datacrypto

import (
	"context"
)

type plainTextEncryptor struct{}

func (enc plainTextEncryptor) Encrypt(_ context.Context, v interface{}, raw *EncryptedRaw) error {
	if raw == nil {
		return newEncryptionError("raw data is nil")
	} else if raw.Alg != AlgPlain {
		return newEncryptionError("unsupported algorithm: %v, expect %v", raw.Alg, AlgPlain)
	}

	dataVersionCorrection(raw)
	switch raw.Ver {
	case V2:
		raw.Raw = v
	default:
		return ErrUnsupportedVersion
	}
	return nil
}

func (enc plainTextEncryptor) Decrypt(_ context.Context, raw *EncryptedRaw, dest interface{}) error {
	if raw == nil {
		return newDecryptionError("raw data is nil")
	}

	switch raw.Ver {
	case V1, V2:
		if raw.Alg != AlgPlain {
			return newDecryptionError("unsupported algorithm: %v, expect %v", raw.Alg, AlgPlain)
		}
		if e := tryAssign(raw.Raw, dest); e != nil {
			return e
		}
	default:
		return ErrUnsupportedVersion
	}
	return nil
}



