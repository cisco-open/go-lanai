package pqcrypt

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/vault"
	"encoding/json"
	"fmt"
	"strconv"
)

// vaultEncryptor implements Encryptor and KeyOperations
type vaultEncryptor struct {
	transit vault.TransitEngine
	props   *KeyProperties
}

func newVaultEncryptor(client *vault.Client, props *KeyProperties) Encryptor {
	return &vaultEncryptor{
		transit: vault.NewTransitEngine(client, func(opt *vault.KeyOption) {
			opt.KeyType = props.Type
			opt.Exportable = props.Exportable
			opt.AllowPlaintextBackup = props.AllowPlaintextBackup
		}),
		props: props,
	}
}

func (enc *vaultEncryptor) Encrypt(ctx context.Context, kid string, v interface{}) (raw *EncryptedRaw, err error) {
	raw = &EncryptedRaw{
		Ver:   V2,
		KeyID: normalizeKeyID(kid),
		Alg:   AlgVault,
	}
	switch {
	case raw.KeyID == "":
		return nil, newEncryptionError("KeyID is required for algorithm %v", raw.Alg)
	}

	if v == nil {
		// special rule encrypted []byte(nil) <-> nil
		return raw, nil
	}

	jsonVal, e := json.Marshal(v)
	if e != nil {
		return nil, newEncryptionError("failed to marshal data - %v", e)
	}
	cipher, e := enc.transit.Encrypt(ctx, raw.KeyID, jsonVal)
	if e != nil {
		return nil, newEncryptionError("encryption engine - %v", e)
	}
	raw.Raw = json.RawMessage(strconv.Quote(string(cipher)))
	return
}

func (enc *vaultEncryptor) Decrypt(ctx context.Context, raw *EncryptedRaw, dest interface{}) error {
	switch {
	case raw == nil:
		return newDecryptionError("raw data is nil")
	case raw.Alg != AlgVault:
		return ErrUnsupportedAlgorithm
	case raw.KeyID == "":
		return newDecryptionError("KeyID is required for algorithm %v", raw.Alg)
	}

	switch raw.Ver {
	case V1, V2:
		return enc.decrypt(ctx, raw, dest)
	default:
		return ErrUnsupportedVersion
	}
}

func (enc *vaultEncryptor) KeyOperations() KeyOperations {
	return enc
}

/* KeyOperations */

func (enc *vaultEncryptor) Create(ctx context.Context, kid string, _ ...KeyOptions) error {
	kid = normalizeKeyID(kid)
	if kid == "" {
		return fmt.Errorf("invalid key ID")
	}
	return enc.transit.PrepareKey(ctx, kid)
}

/* Helpers */

func (enc *vaultEncryptor) decrypt(ctx context.Context, raw *EncryptedRaw, dest interface{}) error {
	if len(raw.Raw) == 0 {
		// special rule encrypted []byte(nil) <-> nil
		return tryAssign(nil, dest)
	}

	var cipher string
	if e := json.Unmarshal(raw.Raw, &cipher); e != nil {
		return newDecryptionError("invalid ciphertext - %v", e)
	}

	plain, e := enc.transit.Decrypt(ctx, normalizeKeyID(raw.KeyID), []byte(cipher))
	if e != nil {
		return newDecryptionError("encryption engine - %v", e)
	}

	switch raw.Ver {
	case V1:
		v, e := extractV1DecryptedPayload(plain)
		if e != nil {
			return newDecryptionError("malformed V1 data - %v", e)
		}
		if e := json.Unmarshal(v, dest); e != nil {
			return newDecryptionError("failed to unmarshal decrypted data - %v", e)
		}
	case V2:
		if e := json.Unmarshal(plain, dest); e != nil {
			return newDecryptionError("failed to unmarshal decrypted data - %v", e)
		}
	}
	return nil
}