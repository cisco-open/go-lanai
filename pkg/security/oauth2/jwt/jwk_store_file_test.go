package jwt

import (
	"context"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rsa"
	"github.com/cisco-open/go-lanai/pkg/security/oauth2"
	"github.com/cisco-open/go-lanai/pkg/utils"
	"github.com/cisco-open/go-lanai/test"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"path/filepath"
	"testing"
	"time"
)

/*************************
	Test Setup
 *************************/

const (
	DefaultPEMKeyID       = "test-key"
	DefaultPEMKeyPassword = "TheCakeIsALie"
	DefaultPEMKeyName     = "test-name"
)

var basicClaims = oauth2.BasicClaims{
	ExpiresAt: time.Now().Add(120 * time.Minute),
	IssuedAt:  time.Now(),
	Issuer:    "test",
	NotBefore: time.Now(),
	Subject:   "test-user",
	ClientId:  "test-client",
}

/*************************
	Test Cases
 *************************/

func TestFileJwkStore(t *testing.T) {
	test.RunTest(context.Background(), t,
		test.GomegaSubTest(SubTestMultiBlockAsymmetricPEM("rsa-2048", false), "RSA-2048"),
		test.GomegaSubTest(SubTestMultiBlockAsymmetricPEM("ec-p256", false), "ECDSA-P-256"),
		test.GomegaSubTest(SubTestMultiBlockAsymmetricPEM("ec-p384", false), "ECDSA-P-384"),
		test.GomegaSubTest(SubTestMultiBlockAsymmetricPEM("ec-p521", false), "ECDSA-P-521"),

		test.GomegaSubTest(SubTestMultiBlockSymmetricPem("hmac-256"), "HMAC-256"),
		test.GomegaSubTest(SubTestMultiBlockSymmetricPem("hmac-384"), "HMAC-384"),
		test.GomegaSubTest(SubTestMultiBlockSymmetricPem("hmac-512"), "HMAC-512"),

		// For ed25519, openssl doesn't have a command to generate private key (either encrypted or unencrypted) in "traditional" format.
		// It can only generate an encrypted private key in pkcs8 format. However, golang standard library doesn't support
		// decoding pem with pkcs8 encrypted format. Therefore, we don't have a test case for encrypted ed25519 private key
		test.GomegaSubTest(SubTestMultiBlockAsymmetricPEM("ed25519", true), "ED25519"),
		test.GomegaSubTest(SubTestSingleBlockPem("hmac-256"), "single-key"),
	)
}

/*************************
	Sub-Test Cases
 *************************/

func SubTestMultiBlockSymmetricPem(prefix string) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		const b = 3
		cases := []TestPEMKey{
			{Name: "SymmetricKeyWithID", Prefix: prefix, File: prefix + ".pem", ID: true, Expect: PrivateBlocks(b)},
			{Name: "SymmetricKeyWithoutID", Prefix: prefix, File: prefix + ".pem", ID: false, Expect: PrivateBlocks(b)},
		}
		opts := make([]test.Options, len(cases))
		for i := range cases {
			opts[i] = test.GomegaSubTest(SubTestLoadPEM(cases[i]), cases[i].Name)
		}
		test.RunTest(ctx, t, opts...)
	}
}

func SubTestMultiBlockAsymmetricPEM(prefix string, skipPasswd bool) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		const b = 3
		cases := []TestPEMKey{
			// with ID
			{Name: "PrivateKeyWithID", Prefix: prefix, File: prefix + "-priv-key.pem", ID: true, Expect: PrivateBlocks(b)},
			{Name: "PrivateKeyTraditionalWithID", Prefix: prefix, File: prefix + "-priv-key-trad.pem", ID: true, Expect: PrivateBlocks(b)},
			{Name: "PublicKeyWithID", Prefix: prefix, File: prefix + "-pub-key.pem", ID: true, Expect: PublicBlocks(b)},
			{Name: "CertificateWithID", Prefix: prefix, File: prefix + "-cert.pem", ID: true, Expect: PublicBlocks(b)},
			// without ID
			{Name: "PrivateKeyWithoutID", Prefix: prefix, File: prefix + "-priv-key.pem", ID: false, Expect: PrivateBlocks(b)},
			{Name: "PrivateKeyTraditionalWithoutID", Prefix: prefix, File: prefix + "-priv-key-trad.pem", ID: true, Expect: PrivateBlocks(b)},
			{Name: "PublicKeyWithoutID", Prefix: prefix, File: prefix + "-pub-key.pem", ID: false, Expect: PublicBlocks(b)},
			{Name: "CertificateWithoutID", Prefix: prefix, File: prefix + "-cert.pem", ID: false, Expect: PublicBlocks(b)},
		}
		if !skipPasswd {
			// with password (encrypted)
			cases = append(cases,
				TestPEMKey{Name: "EncryptedPrivateKeyWithPasswd", Prefix: prefix, File: prefix + "-priv-key-aes256.pem", ID: true, Passwd: true, Expect: PrivateBlocks(b)},
				TestPEMKey{Name: "EncryptedPrivateKeyWithWrongPasswd", Prefix: prefix, File: prefix + "-priv-key-aes256-bad.pem", ID: true, Passwd: true, Expect: ExpectError()},
			)
		}

		opts := make([]test.Options, len(cases))
		for i := range cases {
			opts[i] = test.GomegaSubTest(SubTestLoadPEM(cases[i]), cases[i].Name)
		}
		test.RunTest(ctx, t, opts...)
	}
}

func SubTestSingleBlockPem(prefix string) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		const b = 1
		cases := []TestPEMKey{
			{Name: "SingleKeyWithID", Prefix: prefix, File: prefix + "-single-key.pem", ID: true, Expect: PrivateBlocks(b)},
			{Name: "SingleKeyWithoutID", Prefix: prefix, File: prefix + "-single-key.pem", ID: false, Expect: PrivateBlocks(b)},
		}
		opts := make([]test.Options, len(cases))
		for i := range cases {
			opts[i] = test.GomegaSubTest(SubTestLoadPEM(cases[i]), cases[i].Name)
		}
		test.RunTest(ctx, t, opts...)
	}
}

func SubTestLoadPEM(src TestPEMKey) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		props := MakePEMCryptoProperties(src)
		store := NewFileJwkStore(props)
		if src.Expect.Error {
			AssertEmptyFileJwkStore(ctx, g, store, src)
		} else {
			AssertLoadedFileJwkStore(ctx, g, store, props, src)
		}
	}
}

/*************************
	Helpers
 *************************/

type TestPEMKey struct {
	Prefix string
	Name   string
	File   string
	ID     bool
	Passwd bool
	Expect Expectation
}

type Expectation struct {
	Error   bool
	Blocks  int
	Private int
}

func PublicBlocks(total int) Expectation {
	return Expectation{
		Blocks: total,
	}
}

func PrivateBlocks(total int) Expectation {
	return Expectation{
		Blocks:  total,
		Private: total,
	}
}

func ExpectError() Expectation {
	return Expectation{
		Error: true,
	}
}

func MakePEMCryptoProperties(key TestPEMKey) CryptoProperties {
	return CryptoProperties{
		Keys: map[string]CryptoKeyProperties{
			DefaultPEMKeyName: MakePEMCryptoKeyProperties(key),
		},
	}
}

func MakePEMCryptoKeyProperties(key TestPEMKey) CryptoKeyProperties {
	v := CryptoKeyProperties{
		KeyFormat: string(KeyFileFormatPem),
		Location:  filepath.Join("testdata", key.File),
	}
	if key.ID {
		v.Id = DefaultPEMKeyID
	}
	if key.Passwd {
		v.Password = DefaultPEMKeyPassword
	}
	return v
}

func AssertLoadedFileJwkStore(ctx context.Context, g *gomega.WithT, store *FileJwkStore, props CryptoProperties, src TestPEMKey) {
	// LoadAll
	kids := utils.NewStringSet()
	all, e := store.LoadAll(ctx, DefaultPEMKeyName)
	g.Expect(e).To(Succeed(), "LoadAll should not fail")
	g.Expect(all).To(HaveLen(src.Expect.Blocks), `LoadAll should return correct number of JWKs`)
	var privCount int
	for _, jwk := range all {
		AssertJwkType(ctx, g, src.Prefix, jwk)
		kids.Add(jwk.Id())
		if _, ok := jwk.(PrivateJwk); ok {
			privCount++
		}
		g.Expect(jwk.Name()).To(Equal(DefaultPEMKeyName), "JWK's name should be correct")
		if src.Expect.Blocks > 1 {
			if src.ID {
				g.Expect(jwk.Id()).To(HavePrefix(DefaultPEMKeyID+"-"), `kid should have correct format'`)
			} else {
				g.Expect(jwk.Id()).To(HavePrefix(DefaultPEMKeyName+"-"), `kid should have correct format'`)
			}
		} else {
			g.Expect(jwk.Id()).To(Equal(DefaultPEMKeyName), `kid should be the same as name for single block pem`)
		}
	}
	g.Expect(privCount).To(Equal(src.Expect.Private), `Private JWK count should be correct`)
	g.Expect(len(all)).To(Equal(src.Expect.Blocks), `JWK count should be correct`) //this means the public key count is correct as well
	g.Expect(len(all)).To(Equal(len(kids.Values())), "JWK IDs should be unique")

	// LoadByKid
	kid := all[0].Id()
	jwk, e := store.LoadByKid(ctx, kid)
	g.Expect(e).To(Succeed(), `LoadByKid should not fail`)
	g.Expect(jwk).ToNot(BeZero(), `LoadByKid should return zero value`)

	// LoadByName
	jwk, e = store.LoadByName(ctx, DefaultPEMKeyName)
	g.Expect(e).To(Succeed(), `LoadByName should not fail`)
	g.Expect(jwk).ToNot(BeZero(), `LoadByName should return zero value`)

	// Rotate
	e = store.Rotate(ctx, DefaultPEMKeyName)
	g.Expect(e).To(Succeed(), `Rotate should not fail`)
	next, e := store.LoadByName(ctx, DefaultPEMKeyName)
	g.Expect(e).To(Succeed(), `LoadByName should not fail after Rotate`)
	if src.Expect.Blocks > 1 {
		g.Expect(jwk.Id()).ToNot(Equal(next.Id()), `LoadByName should return different JWK after Rotate`)
	} else {
		g.Expect(jwk.Id()).To(Equal(next.Id()), `LoadByName should return same JWK after Rotate`)
	}

	// assert loaded JWKs with JwtEncoder if applicable (to make sure the loaded JWK has recognized types)
	if privCount > 0 {
		AssertLoadedJwks(ctx, g, store)
	}
}

func AssertEmptyFileJwkStore(ctx context.Context, g *gomega.WithT, store *FileJwkStore, _ TestPEMKey) {
	all, e := store.LoadAll(ctx)
	g.Expect(e).To(Succeed(), `[empty key store] LoadAll should not fail`)
	g.Expect(all).To(BeEmpty(), `[empty key store] LoadAll should return no JWKs`)

	_, e = store.LoadByKid(ctx, DefaultPEMKeyID+`-0`)
	g.Expect(e).To(HaveOccurred(), "[empty key store] LoadByKid should fail")
	_, e = store.LoadByName(ctx, DefaultPEMKeyName)
	g.Expect(e).To(HaveOccurred(), "[empty key store] LoadByName should fail")
	e = store.Rotate(ctx, DefaultPEMKeyName)
	g.Expect(e).To(HaveOccurred(), `[empty key store] Rotate should fail`)
}

func AssertLoadedJwks(ctx context.Context, g *gomega.WithT, store *FileJwkStore) {
	encoder := NewSignedJwtEncoder(SignWithJwkStore(store, DefaultPEMKeyName), SignWithMethod(nil))
	t, e := encoder.Encode(ctx, basicClaims)
	g.Expect(e).To(Succeed(), "using loaded JWK to encode JWT should not fail")
	g.Expect(t).To(Not(BeEmpty()), "using loaded JWK to encode JWT should return empty string")
}

func AssertJwkType(_ context.Context, g *gomega.WithT, expectedType string, jwk Jwk) {
	pubKey := jwk.Public()
	switch expectedType {
	case "ec-p256", "ec-p384", "ec-p521":
		_, ok := pubKey.(*ecdsa.PublicKey)
		g.Expect(ok).To(BeTrue())
	case "ed25519":
		_, ok := pubKey.(ed25519.PublicKey)
		g.Expect(ok).To(BeTrue())
	case "hmac-256", "hmac-384", "hmac-512":
		_, ok := pubKey.([]byte)
		g.Expect(ok).To(BeTrue())
	case "rsa-2048":
		_, ok := pubKey.(*rsa.PublicKey)
		g.Expect(ok).To(BeTrue())
	default:
		g.Fail("unexpected file prefix")
	}
}
