package auth

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/matcher"
	"fmt"
	"github.com/bmatcuk/doublestar/v4"
	"net/url"
	"regexp"
	"strings"
)

const (
	pScheme   = `(?P<scheme>[a-z][a-z0-9+\-.]*)`
	pUserInfo = `(?P<userinfo>[^@]*)`
	pDomain   = `(?P<domain>([a-zA-Z0-9_\-*?]+\.)*([a-zA-Z0-9_\-*?]{1,11}))`
	pPort     = `(?P<port>[0-9*?]{1,5})`
	pPath     = `(?P<path>\/?[^?#:]*)`
	pParams   = `(?P<params>[^#]*)`
	pFragment = `(?P<fragment>.*)`
)

var (
	// Warning: if pattern with custom scheme is provided, it's required to add "/" after ":".
	// 			e.g. "custom-scheme:/some_path" is a valid pattern, but "custom-scheme:some_path" is not
	pUrl = fmt.Sprintf(`^(%s:[/]{1,2})?(%s@)?(%s(:%s)?)?%s(\?%s)?(#%s)?`,
		pScheme, pUserInfo, pDomain, pPort, pPath, pParams, pFragment)

	regexWildcardPattern    = regexp.MustCompile(pUrl)
	regexQueryParamsPattern = regexp.MustCompile(`(?P<key>[^&=]+)(?P<eq>=?)(?P<value>[^&]+)?`)

)

/*****************************
	Public
 *****************************/
// wildcardUrlMatcher implements matcher.Matcher, matcher.ChainableMatcher and fmt.Stringer
// it accept escaped URL string and matches with the defined pattern allowing wildcard * and ? in
// domain, port, and path
type wildcardUrlMatcher struct {
	raw string
	patterns
}

type patterns struct {
	scheme   string
	userInfo string
	domain   string
	port     string
	path     string
	params   map[string][]string
	fragment string
}

// NewWildcardUrlMatcher construct a wildcard URL matcher with given pattern
// The pattern should be escaped for URL endoding
func NewWildcardUrlMatcher(pattern string) (*wildcardUrlMatcher, error) {
	m := wildcardUrlMatcher{
		raw: pattern,
	}
	if e := parsePatterns(pattern, &m.patterns); e != nil {
		return nil, e
	}
	return &m, nil
}

func (m *wildcardUrlMatcher) Matches(i interface{}) (bool, error) {
	switch i.(type) {
	case string:
		return m.urlMatches(i.(string))
	default:
		return false, fmt.Errorf("unsupported URL with type [%T]", i)
	}
}

func (m *wildcardUrlMatcher) MatchesWithContext(_ context.Context, i interface{}) (bool, error) {
	return m.Matches(i)
}

func (m *wildcardUrlMatcher) Or(matchers ...matcher.Matcher) matcher.ChainableMatcher {
	return matcher.Or(m, matchers...)
}

func (m *wildcardUrlMatcher) And(matchers ...matcher.Matcher) matcher.ChainableMatcher {
	return matcher.And(m, matchers...)
}

func (m *wildcardUrlMatcher) String() string {
	return fmt.Sprintf("matches pattern %s", m.raw)
}

/*****************************
	Helpers
 *****************************/
func (m *wildcardUrlMatcher) urlMatches(urlStr string) (bool, error) {
	url, e := url.Parse(urlStr)
	if e != nil {
		return false, e
	}

	// if scheme is given and we cannot map it to valid port, we consider it as custom scheme
	if url.Scheme != "" && schemeToPort(url.Scheme) == "" {
		// for custom scheme, we perform exact match
		return exactMatches(urlStr, m.raw, true), nil
	}

	// exact matches
	mScheme := exactMatches(url.Scheme, m.scheme, false)
	mUserInfo := exactMatches(url.User, m.userInfo, false)
	mQuery := queryParamsMatches(url.Query(), m.params)

	// wildcard matches
	mHost := hostMatches(url.Hostname(), m.domain)
	mPort := portMatches(url.Scheme, url.Port(), m.scheme, m.port, m.domain)
	mPath := pathMatches(url.Path, m.path)

	return mScheme && mUserInfo && mQuery && mHost && mPort && mPath, nil
}

func parsePatterns(raw string, dst *patterns) error {
	// parse overall pattern
	matches := regexWildcardPattern.FindStringSubmatch(raw);
	if matches == nil {
		return fmt.Errorf("invalid pattern %s", raw)
	}

	components := map[string]string{}
	for i, group := range regexWildcardPattern.SubexpNames() {
		components[group] = strings.TrimSpace(matches[i])
	}

	dst.scheme = components["scheme"]
	dst.userInfo = components["userinfo"]
	dst.domain = components["domain"]
	dst.port = components["port"]
	dst.path = components["path"]
	dst.fragment = components["fragment"]

	// parse query params
	dst.params = map[string][]string{}
	all := regexQueryParamsPattern.FindAllStringSubmatch(components["params"], -1)
	if all == nil {
		// no params to parse, ok
		return nil
	}

	for _, one := range all {
		for i, group := range regexQueryParamsPattern.SubexpNames() {
			dst.params[group] = append(dst.params[group], strings.TrimSpace(one[i]))
		}
	}
	return nil
}

// exactMatches check if string representation of given value exactly matches pattern.
// If pattern is empty:
// 1. ignore value and return true, if required == false
// 2. return true only when value is empty, if required == true
// accepted value are string or *url.UserInfo
func exactMatches(value interface{}, pattern string, required bool) bool {
	actual := ""
	switch value.(type) {
	case nil:
		// empty string
	case string:
		actual = value.(string)
	case *url.Userinfo:
		actual = value.(*url.Userinfo).String()
	default:
		return false
	}

	return (pattern == "" && !required) ||
		actual == "" && pattern == "" ||
		pattern != "" && pattern == actual
}

// queryParamsMatches checks whether the pattern query params key and values contains match the actual set
// The actual query params are allowed to contain additional params which will be retained
func queryParamsMatches(query url.Values, pattern map[string][]string) (ret bool) {
	if query == nil {
		query = url.Values{}
	}

	ret = true
	for k, expected := range pattern {
		actual, ok := query[k]
		if !ok {
			continue
		}
		ret = sliceEquals(actual, expected)
		if !ret {
			return
		}
	}
	return
}

// hostMatches host matches allows sub domain to match as well.
// the function returns true if pattern is not set (empty string)
func hostMatches(value, pattern string) bool {
	if !hasWildcard(pattern) {
		return pattern == "" || pattern == value || strings.HasSuffix(value, "." + pattern)
	}
	return wildcardMatches(value, pattern, true)
}

// Check whether the requested port value matches the expected values.
// Special handling is required since the scheme of a expected values is optional in this implementation.
//
// When the patterns for the expected url does not specify a port value, the port value is
// inferred from the scheme of the registered redirect. If that is not specified (which should match
// any scheme):
// 1. 	if domain pattern is specified, the port is inferred based on the scheme of the <EM>requested</EM> URL (which
// 		will match unless the requested URL is using a non-standard port)
// 2. 	if domain pattern is also not set, the port matches any value
//
// The expected patterns may contain a wildcard for the port value.
func portMatches(scheme, port, schemePattern, portPattern, domainPattern string) bool {
	expectedPort := portPattern
	if portPattern == "" {
		switch {
		case schemePattern == "" && domainPattern == "":
			// path-only pattern, match any value
			return true
		case schemePattern == "":
			// domain pattern is specified, port should match scheme
			expectedPort = schemeToPort(scheme)
		default:
			// scheme pattern is specified, use it
			expectedPort = schemeToPort(schemePattern)
		}
	}

	// Implied ports must be made explicit for matching - an empty string will not match a * wildcard!
	if port == "" {
		port = schemeToPort(scheme)
	}

	return wildcardMatches(port, expectedPort, true)

}

// pathMatches matches given path value to pattern with wildcoard support
func pathMatches(value, pattern string) bool {
	if value == "" {
		value = "/"
	}
	if pattern == "" {
		pattern = "/"
	}
	return wildcardMatches(value, pattern, true)
}

// sliceEquals compares two slice and return true if:
// - both slices have same length
// - slice s1 contains all elements of slice s2
func sliceEquals(s1, s2 []string) bool {
	if s1 == nil || s2 == nil || len(s1) != len(s2) {
		return false
	}
	set := utils.NewStringSet(s1...)
	for _, v := range s2 {
		if !set.Has(v) {
			return false
		}
	}
	return true
}

func hasWildcard(pattern string) bool {
	return strings.ContainsAny(pattern, "*?\\[]")
}

// wildcardMatches given string with pattern
// The prefix syntax is:
//
//  prefix:
//    { term }
//  term:
//    '*'         matches any sequence of non-path-separators
//    '**'        matches any sequence of characters, including
//                path separators.
//    '?'         matches any single non-path-separator character
//    '[' [ '^' ] { character-range } ']'
//          character class (must be non-empty)
//    '{' { term } [ ',' { term } ... ] '}'
//    c           matches character c (c != '*', '?', '\\', '[')
//    '\\' c      matches character c
//
//  character-range:
//    c           matches character c (c != '\\', '-', ']')
//    '\\' c      matches character c
//    lo '-' hi   matches character c for lo <= c <= hi
func wildcardMatches(value, pattern string, required bool) bool {
	if pattern == "" {
		return !required || value == ""
	}
	ok, e := doublestar.Match(pattern, value)
	return e == nil && ok
}


func schemeToPort(scheme string) string {
	switch scheme {
	case "http":
		return "80"
	case "https":
		return "443"
	default:
		return ""
	}
}



