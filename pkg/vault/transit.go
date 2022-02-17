package vault

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/hashicorp/vault/api"
	"net/url"
)

const (
	pathTmplCreateKey = `transit/keys/%s`
	pathTmplEncrypt   = `transit/encrypt/%s`
	pathTmplDecrypt   = `transit/decrypt/%s`
)

const (
	defaultTransitKeyType = "aes256-gcm96"
	respKeyCipherText     = "ciphertext"
	respKeyPlainText      = "plaintext"
)

type TransitEngine interface {
	PrepareKey(ctx context.Context, kid string) error
	Encrypt(ctx context.Context, kid string, plaintext []byte) ([]byte, error)
	Decrypt(ctx context.Context, kid string, cipher []byte) ([]byte, error)
}

type KeyOptions func(opt *KeyOption)
type KeyOption struct {
	KeyType              string
	Exportable           bool
	AllowPlaintextBackup bool
}

type transit struct {
	c                *Client
	keyType          string
	exportable       bool
	allowPlaintextBk bool
}

func NewTransitEngine(client *Client, opts ...KeyOptions) TransitEngine {
	opt := KeyOption{
		KeyType: defaultTransitKeyType,
	}
	for _, fn := range opts {
		fn(&opt)
	}
	if opt.KeyType == "" {
		opt.KeyType = defaultTransitKeyType
	}
	return &transit{
		c:                client,
		keyType:          opt.KeyType,
		exportable:       opt.Exportable,
		allowPlaintextBk: opt.AllowPlaintextBackup,
	}
}

func (t *transit) PrepareKey(ctx context.Context, kid string) error {
	path := fmt.Sprintf(pathTmplCreateKey, url.PathEscape(kid))
	req := transitCreateKey{
		Type:                 t.keyType,
		Exportable:           t.exportable,
		AllowPlaintextBackup: t.allowPlaintextBk,
	}

	//nolint:contextcheck
	if _, e := t.c.Logical(ctx).Post(path, &req); e != nil {
		return e
	}
	return nil
}

func (t *transit) Encrypt(ctx context.Context, kid string, plaintext []byte) ([]byte, error) {
	path := fmt.Sprintf(pathTmplEncrypt, url.PathEscape(kid))
	b64 := base64.StdEncoding.EncodeToString(plaintext)
	req := transitEncrypt{
		PlaintextB64: b64,
	}

	s, e := t.c.Logical(ctx).Post(path, &req) //nolint:contextcheck
	if e != nil {
		return nil, e
	}

	ciphertext, e := t.extractString(s, respKeyCipherText)
	return []byte(ciphertext), e
}

func (t *transit) Decrypt(ctx context.Context, kid string, cipher []byte) ([]byte, error) {
	path := fmt.Sprintf(pathTmplDecrypt, url.PathEscape(kid))
	req := transitDecrypt{
		Ciphertext: string(cipher),
	}

	s, e := t.c.Logical(ctx).Post(path, &req) //nolint:contextcheck
	if e != nil {
		return nil, e
	}

	plaintextB64, e := t.extractString(s, respKeyPlainText)
	if e != nil {
		return nil, e
	}
	return base64.StdEncoding.DecodeString(plaintextB64)
}

func (t *transit) post(ctx context.Context, path string, reqData interface{}) (ret *api.Secret, err error) {
	ret, err = t.c.Logical(ctx).Post(path, reqData) //nolint:contextcheck
	switch {
	case err != nil:
		return
	case ret.Data == nil:
		return nil, fmt.Errorf("missing data in vault response")
	}
	return
}

func (t *transit) extractString(s *api.Secret, key string) (string, error) {
	if s.Data == nil {
		return "", fmt.Errorf("missing data in vault response")
	}

	v, ok := s.Data[key]
	if !ok {
		return "", fmt.Errorf("missing %s in vault response data", key)
	}

	text, ok := v.(string)
	if !ok {
		return "", fmt.Errorf("invalid type of %s in vault response data, expected string but got %T", key, v)
	}
	return text, nil
}

/*************************
	Requests
 *************************/

// transitCreateKey is a subset of all supported request parameters of `POST transit/keys/:name`
// see https://www.vaultproject.io/api/secret/transit#create-key
type transitCreateKey struct {
	Type                 string `json:"type"`
	Exportable           bool   `json:"exportable,omitempty"`
	AllowPlaintextBackup bool   `json:"allow_plaintext_backup,omitempty"`
}

// transitEncrypt is a subset of all supported request parameters of `POST /transit/encrypt/:name`
// see https://www.vaultproject.io/api/secret/transit#encrypt-data
type transitEncrypt struct {
	PlaintextB64 string `json:"plaintext"`
	Type         string `json:"type,omitempty"`
	ContextB64   string `json:"context,omitempty"`
	KeyVersion   int    `json:"key_version,omitempty"`
	NonceB64     string `json:"nonce,omitempty"`
}

// transitDecrypt is a subset of all supported request parameters of `POST /transit/decrypt/:name`
// see https://www.vaultproject.io/api/secret/transit#decrypt-data
type transitDecrypt struct {
	Ciphertext string `json:"ciphertext"`
	ContextB64    string `json:"context,omitempty"`
	NonceB64      string `json:"nonce,omitempty"`
}
