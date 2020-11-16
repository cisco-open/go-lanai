package utils

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/matcher"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"fmt"
	"net/http"
	"strings"
)

// TODO support wildcard
// DomainMatcher checks http.Request's Host against expected value with case-insensitive comparison
type DomainMatcher struct {
	expected string
}

func (m *DomainMatcher) Matches(i interface{}) (ret bool, err error) {
	var req *http.Request
	if req, err = interfaceToRequest(i); err != nil {
		return
	}

	return hostMatch(m.expected,req.Host), nil
}

func (m *DomainMatcher) Or(matchers ...matcher.Matcher) matcher.ChainableMatcher {
	return matcher.Or(m, matchers...)
}

func (m *DomainMatcher) And(matchers ...matcher.Matcher) matcher.ChainableMatcher {
	return matcher.And(m, matchers...)
}

func (m *DomainMatcher) String() string {
	return fmt.Sprintf("domain=%s", m.expected)
}

/**************************
	Constructors
***************************/
func WithDomain(expected string) web.MWConditionMatcher {
	return &DomainMatcher{expected: expected}
}

/**************************
	helpers
***************************/
func interfaceToRequest(i interface{}) (*http.Request, error) {
	switch i.(type) {
	case http.Request:
		r := i.(http.Request)
		return &r, nil
	case *http.Request:
		return i.(*http.Request), nil
	default:
		return nil, fmt.Errorf("security condition doesn't support %T", i)
	}
}

func hostMatch(expected, actual string) bool {
	return expected == "" || strings.ToLower(expected) == strings.ToLower(actual)
}
