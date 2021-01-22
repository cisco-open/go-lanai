package oauth2

import (
	"encoding/json"
	"testing"
	. "github.com/onsi/gomega"
	"time"
)
type SubTest func(*testing.T)

var refBasic = map[string]interface{} {
	ClaimAudience: "test-audience",
	ClaimExpire: time.Now(),
	ClaimJwtId: "test-id",
	ClaimIssueAt: time.Now(),
	ClaimIssuer: "test-issuer",
	ClaimNotBefore: time.Now(),
	ClaimSubject: "test-subject",
}

var refEmbedding = map[string]interface{} {
	"first": "test-first-name",
	"last": "test-last-name",
}

var refBasicClaims = BasicClaims{
	Audience: refBasic[ClaimAudience].(string),
	ExpiresAt: refBasic[ClaimExpire].(time.Time),
	Id: refBasic[ClaimJwtId].(string),
	IssuedAt: refBasic[ClaimIssueAt].(time.Time),
	Issuer: refBasic[ClaimIssuer].(string),
	NotBefore: refBasic[ClaimNotBefore].(time.Time),
	Subject: refBasic[ClaimSubject].(string),
}

var refEmbeddingClaims = embeddingClaims{
	BasicClaims: refBasicClaims,
	FirstName: refEmbedding["first"].(string),
	LastName: refEmbedding["last"].(string),
}

/*************************
	Test Cases
 *************************/
func TestBasicClaimsGetters(t *testing.T) {
	claims := &refBasicClaims
	assertClaimsUsingGetter(t, refBasic, claims)
}

func TestBasicClaimsSetters(t *testing.T) {
	// call setters
	claims := &BasicClaims{}
	for k, v := range refBasic {
		claims.Set(k, v)
	}

	assertBasicClaims(t, refBasic, claims)
}

func TestBasicClaimsJson(t *testing.T) {
	claims := &refBasicClaims

	// marshal
	str, err := json.Marshal(claims)
	g := NewWithT(t)
	g.Expect(err).NotTo(HaveOccurred(), "JSON marshal should not return error")
	g.Expect(str).NotTo(BeZero(), "JSON marshal should not return empty string")

	t.Logf("JSON: %s", str)

	// unmarshal
	parsed := BasicClaims{}
	err = json.Unmarshal([]byte(str), &parsed)
	g.Expect(err).NotTo(HaveOccurred(), "JSON unmarshal should not return error")
	assertBasicClaims(t, refBasic, claims)
}

func TestEmbeddingClaimsGetters(t *testing.T) {
	claims := &refEmbeddingClaims
	assertClaimsUsingGetter(t, refBasic, claims)
	assertClaimsUsingGetter(t, refEmbedding, claims)
}

func TestEmbeddingClaimsSetters(t *testing.T) {
	// call setters
	claims := &embeddingClaims{}
	for k, v := range refBasic {
		claims.Set(k, v)
	}

	for k, v := range refEmbedding {
		claims.Set(k, v)
	}

	assertEmbeddingClaims(t, refBasic, claims)
}

func TestEmbeddingClaimsJson(t *testing.T) {
	claims := &refEmbeddingClaims

	// marshal
	str, err := json.Marshal(claims)
	g := NewWithT(t)
	g.Expect(err).NotTo(HaveOccurred(), "JSON marshal should not return error")
	g.Expect(str).NotTo(BeZero(), "JSON marshal should not return empty string")

	t.Logf("JSON: %s", str)

	// unmarshal
	parsed := BasicClaims{}
	err = json.Unmarshal([]byte(str), &parsed)
	g.Expect(err).NotTo(HaveOccurred(), "JSON unmarshal should not return error")
	assertEmbeddingClaims(t, refBasic, claims)
}

/*************************
	Helpers
 *************************/
func assertClaimsUsingGetter(t *testing.T, ref map[string]interface{}, actual Claims) {
	g := NewWithT(t)
	for k, v := range ref {
		g.Expect(actual.Get(k)).To(BeEquivalentTo(v), "claim [%s] should be correct", k)
	}
}

func assertBasicClaims(t *testing.T, ref map[string]interface{}, actual *BasicClaims) {
	g := NewWithT(t)
	g.Expect(actual.Audience).To(Equal(ref[ClaimAudience]))
	g.Expect(actual.ExpiresAt).To(Equal(ref[ClaimExpire]))
	g.Expect(actual.Id).To(Equal(ref[ClaimJwtId]))
	g.Expect(actual.IssuedAt).To(Equal(ref[ClaimIssueAt]))
	g.Expect(actual.Issuer).To(Equal(ref[ClaimIssuer]))
	g.Expect(actual.NotBefore).To(Equal(ref[ClaimNotBefore]))
	g.Expect(actual.Subject).To(Equal(ref[ClaimSubject]))
}

func assertEmbeddingClaims(t *testing.T, ref map[string]interface{}, actual *embeddingClaims) {
	g := NewWithT(t)
	g.Expect(actual.Audience).To(Equal(ref[ClaimAudience]))
	g.Expect(actual.ExpiresAt).To(Equal(ref[ClaimExpire]))
	g.Expect(actual.Id).To(Equal(ref[ClaimJwtId]))
	g.Expect(actual.IssuedAt).To(Equal(ref[ClaimIssueAt]))
	g.Expect(actual.Issuer).To(Equal(ref[ClaimIssuer]))
	g.Expect(actual.NotBefore).To(Equal(ref[ClaimNotBefore]))
	g.Expect(actual.Subject).To(Equal(ref[ClaimSubject]))
	g.Expect(actual.FirstName).To(Equal(refEmbedding["first"]))
	g.Expect(actual.LastName).To(Equal(refEmbedding["last"]))
}

/*************************
	composite Type
 *************************/
type embeddingClaims struct {
	StructClaimsMapper
	BasicClaims
	FirstName  string    `claim:"first"`
	LastName  string    `claim:"last"`
}

func (c *embeddingClaims) MarshalJSON() ([]byte, error) {
	return c.StructClaimsMapper.DoMarshalJSON(c)
}

func (c *embeddingClaims) UnmarshalJSON(bytes []byte) error {
	return c.StructClaimsMapper.DoUnmarshalJSON(c, bytes)
}

func (c *embeddingClaims) Get(claim string) interface{} {
	return c.StructClaimsMapper.Get(c, claim)
}

func (c *embeddingClaims) Has(claim string) bool {
	return c.StructClaimsMapper.Has(c, claim)
}

func (c *embeddingClaims) Set(claim string, value interface{}) {
	c.StructClaimsMapper.Set(c, claim, value)
}


