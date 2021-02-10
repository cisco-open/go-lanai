package auth

import (
	"context"
	"fmt"
	"testing"
	. "github.com/onsi/gomega"
)

/*************************
	Test Cases
 *************************/
func TestPositiveMatch(t *testing.T) {

	// subdomain match
	t.Run("Subdomain", PositiveTest("http://saml.wildcard.com", "http://wildcard.com"))
	t.Run("Subdomain", PositiveTest("http://saml.qa.wildcard.com", "http://wildcard.com"))
	t.Run("Subdomain", PositiveTest("http://saml.qa.wildcard.com", "http://qa.wildcard.com"))

	// exact match
	t.Run("Exact", PositiveTest(
		"http://user:password@test.com:1234/path/to/redirect?param2=val2&param1=val1#fragment",
		"http://user:password@test.com:1234/path/to/redirect?param1=val1&param2=val2") )
	t.Run("Exact No Optional", PositiveTest(
		"http://idp.test.com/path/to/redirect?param2=val2&param1=val1",
		"test.com:80/path/to/redirect?param1=val1&param2=val2") )

	// wildcard match
	t.Run("Wildcard", PositiveTest(
		"https://user:password@wildcard.test.com:4321/path/to/redirect.html?param1=val1&param2=val2",
		"https://user:password@*.test.com:*/**/redirect.*?param2=val2&param1=val1") )

	// wildcard without optional patterns match
	t.Run("Wildcard No Optional", PositiveTest(
		"http://user:password@wildcard.test.com:4321/path/to/redirect.html?param1=val1&param2=val2",
		"*.test.com:*/path/to/redirect.html") )
	t.Run("Wildcard No Optional", PositiveTest(
		"http://user:password@wildcard.test.com/path/to/redirect.html?param1=val1&param2=val2",
		"*.test.com/path/to/redirect.html") )


	// ? test
	t.Run("Single Char", PositiveTest(
		"https://wild.test.com:123/path/to/redirect.html?param1=val1&param2=val2",
		"????.test.com:???/path/**") )

	// wildcard port
	t.Run("NonStdPort", PositiveTest(
		"https://test.com:8443",
		"https://test.com:*") )
	t.Run("StdPort Explicit-Wildcard", PositiveTest(
		"https://test.com:443",
		"https://test.com:*") )
	t.Run("StdPort Implicit-Wildcard", PositiveTest(
		"https://test.com",
		"https://test.com:*") )
	t.Run("StdPort Implicit-Explicit", PositiveTest(
		"http://wildcard.test.com",
		"http://wildcard.test.com:80") )
	t.Run("StdPort Explicit-Implicit", PositiveTest(
		"https://wildcard.test.com:443",
		"https://wildcard.test.com") )
	t.Run("StdPort Implicit-Explicit", PositiveTest(
		"https://wildcard.test.com",
		"https://wildcard.test.com:443") )

	// normalized path
	t.Run("Normialized SsoPath", PositiveTest(
		"https://wildcard.test.com",
		"https://wildcard.test.com/") )
	t.Run("Normialized SsoPath", PositiveTest(
		"https://wildcard.test.com/",
		"https://wildcard.test.com") )
	t.Run("Normialized SsoPath", PositiveTest(
		"http://wildcard.test.com:80",
		"http://wildcard.test.com/") )

	// custom scheme
	t.Run("Custom Scheme", PositiveTest(
		"com.cisco.msx:/oauth_call_back",
		"com.cisco.msx:/oauth_call_back") )

	//t.Run("", PositiveTest(
	//	"",
	//	"") )
}

func TestNegativeExactMatch(t *testing.T) {

	// scheme mismatch
	t.Run("Scheme", NegativeTest(
		"https://wildcard.test.com:1234/path/to/redirect.html?param1=val1&param2=val2",
		"http://*.test.com:*/**") )

	t.Run("Scheme", NegativeTest(
		"https://wildcard.test.com:1234",
		"ftp://*.test.com:*/") )

	// user info mismatch
	t.Run("UserInfo", NegativeTest(
		"http://user@wildcard.test.com:1234/path/to/redirect.html?param1=val1&param2=val2",
		"http://user:password@*.test.com:*/**"))
	t.Run("UserInfo", NegativeTest(
		"http://wildcard.test.com:1234/path/to/redirect.html?param1=val1&param2=val2",
		"http://user:password@*.test.com:*/**"))

	// Implied ports don't match
	t.Run("Implied Port", NegativeTest(
		"http://wildcard.test.com/",
		"https://wildcard.test.com/") )

	// params mismatch
	t.Run("Query", NegativeTest(
		"https://wildcard.test.com:1234/path/to/redirect.html?param1=val1",
		"http://*.test.com:*/**?param1=val1&param2=val2") )
	t.Run("Query", NegativeTest(
		"https://wildcard.test.com:1234/path/to/redirect.html?param1=val1&param2=derp",
		"http://*.test.com:*/**?param1=val1&param2=val2") )
	t.Run("Query", NegativeTest(
		"https://wildcard.test.com:1234/path/to/redirect.html?param1=val1&param2=val2&param1=derp",
		"http://*.test.com:*/**?param1=val1&param2=val2") )

	// path mismatch
	t.Run("SsoPath", NegativeTest(
		"https://wildcard.test.com:4321/path/to/redirect.html?param1=val1&param2=val2",
		"*.test.com:*/path/") )
	t.Run("SsoPath", NegativeTest(
		"https://wildcard.test.com/path/to/redirect.html?param1=val1&param2=val2",
		"wildcard.test.com/") )

	// Domain
	t.Run("Domain", NegativeTest(
		"http://samlwildcard.com",
		"http://wildcard.com") )
	t.Run("Domain", NegativeTest(
		"http://saml.wildcard.msx.com",
		"http://wildcard.com") )
	t.Run("Domain", NegativeTest(
		"http://saml.wildcard",
		"http://wildcard.com") )
	//t.Run("", NegativeTest("", "") )
}

func TestNegativeWildcardMatch(t *testing.T) {

	// host mismatch
	t.Run("Host MutliChar", NegativeTest(
		"https://derp.test.com:1234/path/to/redirect.html?param1=val1&param2=val2",
		"http://wild*.test.com:*/**") )
	t.Run("Host SingleChar", NegativeTest(
		"https://derpina.test.com:123/path/to/redirect.html?param1=val1&param2=val2",
		"http://????.test.com:*/**") )

	// port mismatch
	t.Run("Port MultiChar", NegativeTest(
		"https://wildcard.test.com:2345/path/to/redirect.html?param1=val1&param2=val2",
		"http://wildcard.test.com:1*/**") )
	t.Run("Port SingleChar", NegativeTest(
		"https://wildcard.test.com:2345/path/to/redirect.html?param1=val1&param2=val2",
		"http://*.test.com:???/**") )
	t.Run("Port Implicit", NegativeTest(
		"https://wildcard.test.com:2345/path/to/redirect.html?param1=val1&param2=val2",
		"http://*.test.com/**") )

	// path mismatch
	t.Run("SsoPath Exact", NegativeTest(
		"https://wildcard.test.com:1234/not/path/to/redirect.html?param1=val1&param2=val2",
		"http://*.test.com:*/path/to/redirect.html") )
	t.Run("SsoPath MultiChar", NegativeTest(
		"https://wildcard.test.com:1234/path/to/redirect.jsp?param1=val1&param2=val2",
		"http://*.test.com:*/path/to/*.html") )
	t.Run("SsoPath NoPattern", NegativeTest(
		"https://wildcard.test.com",
		"http://*.test.com:*/") )
	//t.Run("", NegativeTest(
	//	"",
	//	"") )
}

func TestInvalidPattern(t *testing.T) {
	t.Run("Scheme Wildcard", NegativeTest(
		"https://wildcard.test.com:1234",
		"*://*.test.com:*/"))
	t.Run("UserInfo Wildcard", NegativeTest(
		"http://user@wildcard.test.com:*",
		"http://*@*.test.com:*/"))
	t.Run("Query Wildcard", NegativeTest(
		"http://user@wildcard.test.com:*?param=abc",
		"http://*@*.test.com:*/?param=*"))
	//t.Run("", InvalidPatternTest(""))
}

/*************************
	Sub Tests
 *************************/
func PositiveTest(actual, pattern string) func(*testing.T) {
	return func(t *testing.T) {
		matcher, e := NewWildcardUrlMatcher(pattern)

		g := NewWithT(t)
		g.Expect(e).To(Succeed(), `"%s" should be a valid pattern`, pattern)

		ctx := context.Background()
		desc := fmt.Sprintf(`"%s" should match pattern "%s"`, actual, pattern)
		g.Expect(matcher.Matches(actual)).To(BeTrue(), desc+ " with out context")
		g.Expect(matcher.MatchesWithContext(ctx, actual)).To(BeTrue(), desc+ " with context")
	}
}

func NegativeTest(actual, pattern string) func(*testing.T) {
	return func(t *testing.T) {
		matcher, e := NewWildcardUrlMatcher(pattern)

		g := NewWithT(t)
		g.Expect(e).To(Succeed(), `"%s" should be a valid pattern`, pattern)

		ctx := context.Background()
		desc := fmt.Sprintf(`"%s" should not match pattern "%s"`, actual, pattern)

		if ret, e := matcher.Matches(actual); e != nil {
			g.Expect(ret).To(BeFalse(), desc + " with out context")
		}
		if ret, e := matcher.MatchesWithContext(ctx, actual); e != nil {
			g.Expect(ret).To(BeFalse(), desc + " with context")
		}
	}
}

/*************************
	Helpers
 *************************/
