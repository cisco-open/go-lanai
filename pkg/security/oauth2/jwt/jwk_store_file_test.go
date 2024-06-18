package jwt

import (
	"context"
	"github.com/cisco-open/go-lanai/pkg/security/oauth2"
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
	ExpiresAt:         time.Now().Add(120 * time.Minute),
	IssuedAt:          time.Now(),
	Issuer:            "test",
	NotBefore:         time.Now(),
	Subject:           "test-user",
	ClientId:          "test-client",
}

/*************************
	Test Cases
 *************************/

func TestFileJwkStore(t *testing.T) {
	test.RunTest(context.Background(), t,
		test.GomegaSubTest(SubTestMultiBlockPEM("rsa-2048", false), "RSA-2048"),
		// TODO ECDSA should be supported
		//test.GomegaSubTest(SubTestMultiBlockPEM("ec-p256", false), "ECDSA-P-256"),
		//test.GomegaSubTest(SubTestMultiBlockPEM("ec-p384", false), "ECDSA-P-384"),
		//test.GomegaSubTest(SubTestMultiBlockPEM("ec-p521", false), "ECDSA-P-521"),
		// TODO more tests for MAC secret
		// TODO ed25519 should be supported
		//test.GomegaSubTest(SubTestMultiBlockPEM("ed25519", true), "ED25519"),
		// TODO find a way to decrypt password protected private key (any kind) without using x509.DecryptPEMBlock
	)

}

/*************************
	Sub-Test Cases
 *************************/

func SubTestMultiBlockPEM(prefix string, skipPasswd bool) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		const b = 3
		cases := []TestPEMKey{
			// with ID
			{Name: "PrivateKeyWithID", File: prefix + "-priv-key.pem", ID: true, Expect: PrivateBlocks(b)},
			{Name: "PublicKeyWithID", File: prefix + "-pub-key.pem", ID: true, Expect: PublicBlocks(b)},
			{Name: "CertificateWithID", File: prefix + "-cert.pem", ID: true, Expect: PublicBlocks(b)},
			// without ID
			{Name: "PrivateKeyWithoutID", File: prefix + "-priv-key.pem", ID: false, Expect: PrivateBlocks(b)},
			{Name: "PublicKeyWithoutID", File: prefix + "-pub-key.pem", ID: false, Expect: PublicBlocks(b)},
			{Name: "CertificateWithoutID", File: prefix + "-cert.pem", ID: false, Expect: PublicBlocks(b)},
		}
		if !skipPasswd {
			// with password (encrypted)
			cases = append(cases,
				TestPEMKey{Name: "EncryptedPrivateKeyWithPasswd", File: prefix + "-priv-key-aes256.pem", ID: true, Passwd: true, Expect: PrivateBlocks(b)},
				TestPEMKey{Name: "EncryptedPrivateKeyWithWrongPasswd", File: prefix + "-priv-key-aes256-bad.pem", ID: true, Passwd: true, Expect: ExpectError()},
			)
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
			AssertLoadedFileJwkStore(ctx, g, store, src)
		}
	}
}

/*************************
	Helpers
 *************************/

type TestPEMKey struct {
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

func AssertLoadedFileJwkStore(ctx context.Context, g *gomega.WithT, store *FileJwkStore, src TestPEMKey) {
	// LoadAll
	all, e := store.LoadAll(ctx, DefaultPEMKeyName)
	g.Expect(e).To(Succeed(), "LoadAll should not fail")
	g.Expect(all).To(HaveLen(src.Expect.Blocks), `LoadAll should return correct number of JWKs`)
	var privCount int
	for _, jwk := range all {
		if _, ok := jwk.(PrivateJwk); ok {
			privCount++
		}
		g.Expect(jwk.Name()).To(Equal(DefaultPEMKeyName), "JWK's name should be correct")
		if src.ID {
			g.Expect(jwk.Id()).To(HavePrefix(DefaultPEMKeyID+"-"), `kid should have correct format'`)
		} else {
			g.Expect(jwk.Id()).To(HavePrefix(DefaultPEMKeyName+"-"), `kid should have correct format'`)
		}
	}
	g.Expect(privCount).To(Equal(src.Expect.Private), `Private JWK count should be correct`)

	// LoadByKid
	kid := all[0].Id()
	jwk, e := store.LoadByKid(ctx, kid)
	g.Expect(e).To(Succeed(), `LoadByKid should not fail`)
	g.Expect(jwk).ToNot(BeZero(), `LoadByKid should return zero value`)

	// LoadByKid
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
	encoder := NewSignedJwtEncoder(SignWithJwkStore(store, DefaultPEMKeyName))
	t, e := encoder.Encode(ctx, basicClaims)
	g.Expect(e).To(Succeed(), "using loaded JWK to encode JWT should not fail")
	g.Expect(t).To(Not(BeEmpty()), "using loaded JWK to encode JWT should return empty string")
}
