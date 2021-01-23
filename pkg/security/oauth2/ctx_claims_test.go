package oauth2

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"encoding/json"
	. "github.com/onsi/gomega"
	"testing"
	"time"
)

const ExtraKey = "additional"

var now = time.Now().Truncate(time.Second)

var refBasic = map[string]interface{} {
	ClaimAudience: "test-audience",
	ClaimExpire: now,
	ClaimJwtId: "test-id",
	ClaimIssueAt: now,
	ClaimIssuer: "test-issuer",
	ClaimNotBefore: now,
	ClaimSubject: "test-subject",
	ClaimClientId: "test-client-id",
	ClaimScope: utils.NewStringSet("read", "write"),
}

var refMore = map[string]interface{} {
	ClaimAudience: "test-audience",
	ClaimExpire: now,
	ClaimJwtId: "test-id",
	ClaimIssueAt: now,
	ClaimIssuer: "test-issuer",
	ClaimNotBefore: now,
	ClaimSubject: "test-subject",
	ClaimClientId: "test-client-id",
	ClaimScope: utils.NewStringSet("read", "write"),
	"first": "test-first-name",
	"last": "test-last-name",
}

var refExtra = map[string]interface{} {
	ClaimAudience: "test-audience",
	ClaimExpire: now,
	ClaimJwtId: "test-id",
	ClaimIssueAt: now,
	ClaimIssuer: "test-issuer",
	ClaimNotBefore: now,
	ClaimSubject: "test-subject",
	ClaimClientId: "test-client-id",
	ClaimScope: utils.NewStringSet("read", "write"),
	"first": "test-first-name",
	"last": "test-last-name",
	"extra1": "test-extra-value-1",
	"extra2": "test-extra-value-2",
}



/*************************
	Test Cases
 *************************/
func TestBasicClaims(t *testing.T) {
	t.Run("GettersTest", GettersTest(newRefBasicClaims(), refBasic))
	t.Run("SettersTest", SettersTest(newRefBasicClaims(), refBasic, assertBasicClaims))
	t.Run("JsonTest", JsonTest(newRefBasicClaims(), &BasicClaims{}, refBasic, assertBasicClaims))
}

func TestFieldEmbeddingClaims(t *testing.T) {
	t.Run("GettersTest", GettersTest(newRefFieldEmbeddingClaims(), refMore))
	t.Run("SettersTest", SettersTest(newRefFieldEmbeddingClaims(), refMore, assertFieldEmbeddingClaims))
	t.Run("JsonTest", JsonTest(newRefFieldEmbeddingClaims(), &fieldEmbeddingClaims{}, refMore, assertFieldEmbeddingClaims))
}

func TestInterfaceEmbeddingClaims(t *testing.T) {
	t.Run("GettersTest",
		GettersTest(newRefInterfaceEmbeddingClaims(), refExtra))
	t.Run("SettersTest",
		SettersTest(newRefInterfaceEmbeddingClaims(), refExtra, assertInterfaceEmbeddingClaims))
	t.Run("JsonTest",
		JsonTest(newRefInterfaceEmbeddingClaims(), &interfaceEmbeddingClaims{
			Claims: &BasicClaims{},
			Extra: MapClaims{},
		}, refExtra, assertInterfaceEmbeddingClaims))
	t.Run("JsonTestWithDifferentEmbedded",
		JsonTest(newRefInterfaceEmbeddingClaims(), &interfaceEmbeddingClaims{
			Claims: MapClaims{},
			Extra: MapClaims{},
		}, refExtra, assertAlternativeInterfaceEmbeddingClaims))
}

/*************************
	Sub Tests
 *************************/
func GettersTest(claims Claims, expected map[string]interface{}) func(*testing.T) {
	return func(t *testing.T) {
		assertClaimsUsingGetter(t, expected, claims)
	}
}

func SettersTest(claims Claims, values map[string]interface{}, assertFunc claimsAssertion) func(*testing.T) {
	return func(t *testing.T) {
		// call setters
		for k, v := range values {
			claims.Set(k, v)
		}
		assertFunc(t, values, claims)
	}
}

func JsonTest(claims Claims, empty Claims, expected map[string]interface{}, assertFunc claimsAssertion) func(*testing.T) {
	return func(t *testing.T) {
		// marshal
		str, err := json.Marshal(claims)
		g := NewWithT(t)
		g.Expect(err).NotTo(HaveOccurred(), "JSON marshal should not return error")
		g.Expect(str).NotTo(BeZero(), "JSON marshal should not return empty string")

		t.Logf("JSON: %s", str)

		// unmarshal
		parsed := empty
		err = json.Unmarshal([]byte(str), &parsed)
		g.Expect(err).NotTo(HaveOccurred(), "JSON unmarshal should not return error")

		assertFunc(t, expected, parsed)
	}
}

/*************************
	Helpers
 *************************/
type claimsAssertion func(*testing.T, map[string]interface{}, Claims)

func assertClaimsUsingGetter(t *testing.T, ref map[string]interface{}, actual Claims) {
	g := NewWithT(t)
	for k, v := range ref {
		if k == ExtraKey {
			continue
		}
		g.Expect(actual.Get(k)).To(BeEquivalentTo(v), "claim [%s] should be correct", k)
	}
}

func assertBasicClaims(t *testing.T, ref map[string]interface{}, claims Claims) {
	g := NewWithT(t)
	actual := claims.(*BasicClaims)
	g.Expect(actual.Audience).To(Equal(ref[ClaimAudience]))
	g.Expect(actual.ExpiresAt).To(Equal(ref[ClaimExpire]))
	g.Expect(actual.Id).To(Equal(ref[ClaimJwtId]))
	g.Expect(actual.IssuedAt).To(Equal(ref[ClaimIssueAt]))
	g.Expect(actual.Issuer).To(Equal(ref[ClaimIssuer]))
	g.Expect(actual.NotBefore).To(Equal(ref[ClaimNotBefore]))
	g.Expect(actual.Subject).To(Equal(ref[ClaimSubject]))
	g.Expect(actual.Scopes).To(Equal(ref[ClaimScope]))
	g.Expect(actual.ClientId).To(Equal(ref[ClaimClientId]))
}

func assertFieldEmbeddingClaims(t *testing.T, ref map[string]interface{}, claims Claims) {
	g := NewWithT(t)
	actual := claims.(*fieldEmbeddingClaims)
	g.Expect(actual.Audience).To(Equal(ref[ClaimAudience]))
	g.Expect(actual.ExpiresAt).To(Equal(ref[ClaimExpire]))
	g.Expect(actual.Id).To(Equal(ref[ClaimJwtId]))
	g.Expect(actual.IssuedAt).To(Equal(ref[ClaimIssueAt]))
	g.Expect(actual.Issuer).To(Equal(ref[ClaimIssuer]))
	g.Expect(actual.NotBefore).To(Equal(ref[ClaimNotBefore]))
	g.Expect(actual.Subject).To(Equal(ref[ClaimSubject]))
	g.Expect(actual.Scopes).To(Equal(ref[ClaimScope]))
	g.Expect(actual.ClientId).To(Equal(ref[ClaimClientId]))
	g.Expect(actual.FirstName).To(Equal(ref["first"]))
	g.Expect(actual.LastName).To(Equal(ref["last"]))
}

func assertInterfaceEmbeddingClaims(t *testing.T, ref map[string]interface{}, claims Claims) {
	g := NewWithT(t)
	actual := claims.(*interfaceEmbeddingClaims)
	g.Expect(actual.FirstName).To(Equal(ref["first"]))
	g.Expect(actual.LastName).To(Equal(ref["last"]))

	for k, v := range ref {
		g.Expect(actual.Get(k)).To(Equal(v))
	}
}

func assertAlternativeInterfaceEmbeddingClaims(t *testing.T, ref map[string]interface{}, claims Claims) {
	g := NewWithT(t)
	actual := claims.(*interfaceEmbeddingClaims)
	for k, v := range ref {
		if t, ok := v.(time.Time); ok {
			g.Expect(actual.Get(k)).To(Equal(t.Unix()))
		} else if s, ok := v.(utils.StringSet); ok {
			g.Expect(actual.Get(k)).To(BeEquivalentTo(s.ToSet().Values()))
		} else {
			g.Expect(actual.Get(k)).To(Equal(v))
		}
	}
}

func newRefBasicClaims() *BasicClaims {
	return &BasicClaims {
		Audience: refBasic[ClaimAudience].(string),
		ExpiresAt: refBasic[ClaimExpire].(time.Time),
		Id: refBasic[ClaimJwtId].(string),
		IssuedAt: refBasic[ClaimIssueAt].(time.Time),
		Issuer: refBasic[ClaimIssuer].(string),
		NotBefore: refBasic[ClaimNotBefore].(time.Time),
		Subject: refBasic[ClaimSubject].(string),
		Scopes: refBasic[ClaimScope].(utils.StringSet),
		ClientId: refBasic[ClaimClientId].(string),
	}
}

func newRefFieldEmbeddingClaims() *fieldEmbeddingClaims {
	return &fieldEmbeddingClaims{
		BasicClaims: *newRefBasicClaims(),
		FirstName:   refMore["first"].(string),
		LastName:    refMore["last"].(string),
	}
}

func newRefInterfaceEmbeddingClaims() *interfaceEmbeddingClaims {
	basic := newRefBasicClaims()
	return &interfaceEmbeddingClaims{
		Claims:  basic,
		FirstName: refExtra["first"].(string),
		LastName:  refExtra["last"].(string),
		Extra: MapClaims{
			"extra1": refExtra["extra1"],
			"extra2": refExtra["extra2"],
		},
	}
}

/*************************
	composite Type
 *************************/
// fieldEmbeddingClaims
type fieldEmbeddingClaims struct {
	FieldClaimsMapper
	BasicClaims
	FirstName  string    `claim:"first"`
	LastName  string    `claim:"last"`
}

func (c *fieldEmbeddingClaims) MarshalJSON() ([]byte, error) {
	return c.FieldClaimsMapper.DoMarshalJSON(c)
}

func (c *fieldEmbeddingClaims) UnmarshalJSON(bytes []byte) error {
	return c.FieldClaimsMapper.DoUnmarshalJSON(c, bytes)
}

func (c *fieldEmbeddingClaims) Get(claim string) interface{} {
	return c.FieldClaimsMapper.Get(c, claim)
}

func (c *fieldEmbeddingClaims) Has(claim string) bool {
	return c.FieldClaimsMapper.Has(c, claim)
}

func (c *fieldEmbeddingClaims) Set(claim string, value interface{}) {
	c.FieldClaimsMapper.Set(c, claim, value)
}

// interfaceEmbeddingClaims
type interfaceEmbeddingClaims struct {
	FieldClaimsMapper
	Claims
	FirstName string `claim:"first"`
	LastName  string `claim:"last"`
	Extra     Claims
}

func (c *interfaceEmbeddingClaims) MarshalJSON() ([]byte, error) {
	return c.FieldClaimsMapper.DoMarshalJSON(c)
}

func (c *interfaceEmbeddingClaims) UnmarshalJSON(bytes []byte) error {
	return c.FieldClaimsMapper.DoUnmarshalJSON(c, bytes)
}

func (c *interfaceEmbeddingClaims) Get(claim string) interface{} {
	return c.FieldClaimsMapper.Get(c, claim)
}

func (c *interfaceEmbeddingClaims) Has(claim string) bool {
	return c.FieldClaimsMapper.Has(c, claim)
}

func (c *interfaceEmbeddingClaims) Set(claim string, value interface{}) {
	c.FieldClaimsMapper.Set(c, claim, value)
}

