package pqcrypt

import (
	"context"
	"encoding/json"
)

type plainTextEncryptor struct{}

func (enc plainTextEncryptor) Encrypt(_ context.Context, kid string, v interface{}) (raw *EncryptedRaw, err error) {
	raw = &EncryptedRaw{
		Ver:   V2,
		KeyID: kid,
		Alg:   AlgPlain,
	}
	switch {
	case raw.KeyID == "":
		return nil, newEncryptionError("KeyID is required for algorithm %v", raw.Alg)
	}

	data, e := json.Marshal(v)
	if e != nil {
		return nil, newEncryptionError("cannot marshal data to JSON - %v", e)
	}
	raw.Raw = data
	return
}

func (enc plainTextEncryptor) Decrypt(ctx context.Context, raw *EncryptedRaw, dest interface{}) error {
	if raw == nil {
		return newDecryptionError("raw data is nil")
	}

	switch raw.Ver {
	case V1, V2:
		return enc.decrypt(ctx, raw, dest)
	default:
		return ErrUnsupportedVersion
	}
}

func (enc plainTextEncryptor) KeyOperations() KeyOperations {
	return noopKeyOps
}

func (enc plainTextEncryptor) decrypt(_ context.Context, raw *EncryptedRaw, dest interface{}) error {
	if raw.Alg != AlgPlain {
		return ErrUnsupportedAlgorithm
	}
	switch raw.Ver {
	case V1:
		v, e := extractV1DecryptedPayload(raw.Raw)
		if e != nil {
			return newDecryptionError("malformed V1 data - %v", e)
		}
		if e := json.Unmarshal(v, dest); e != nil {
			return newDecryptionError("failed to unmarshal decrypted data - %v", e)
		}
	case V2:
		if e := json.Unmarshal(raw.Raw, dest); e != nil {
			return newDecryptionError("failed to unmarshal decrypted data - %v", e)
		}
	}
	return nil
}


