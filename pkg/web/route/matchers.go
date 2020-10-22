package route

import (
	"cto-github.cisco.com/livdu/jupiter/pkg/utils/matcher"
	"cto-github.cisco.com/livdu/jupiter/pkg/web"
	"fmt"
	"github.com/bmatcuk/doublestar"
	pathutils "path"
	"regexp"
	"strings"
)

// MethodMatcher checks web.Route's method with case-insensitive comparison
type MethodMatcher struct {
	expected string
}

func (m *MethodMatcher) Matches(i interface{}) (ret bool, err error) {
	var route *web.Route
	if route, err = interfaceToRoute(i); err != nil {
		return
	}
	return methodMatch(m.expected, route.Method), nil
}

func (m *MethodMatcher) Or(matchers ...matcher.Matcher) matcher.ChainableMatcher {
	return matcher.Or(m, matchers...)
}

func (m *MethodMatcher) And(matchers ...matcher.Matcher) matcher.ChainableMatcher {
	return matcher.And(m, matchers...)
}

// PatternMatcher checks web.Route's path with prefix
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
type PatternMatcher struct {
	pattern string
}

func (m *PatternMatcher) Matches(i interface{}) (ret bool, err error) {
	var route *web.Route
	if route, err = interfaceToRoute(i); err != nil {
		return
	}

	path := absPath(route)
	ret, err = pathMatchPattern(m.pattern, path)
	return
}

func (m *PatternMatcher) Or(matchers ...matcher.Matcher) matcher.ChainableMatcher {
	return matcher.Or(m, matchers...)
}

func (m *PatternMatcher) And(matchers ...matcher.Matcher) matcher.ChainableMatcher {
	return matcher.And(m, matchers...)
}

// PatternMatcher checks web.Route's path with given prefix
type PrefixMatcher struct {
	prefix string
}

func (m *PrefixMatcher) Matches(i interface{}) (ret bool, err error) {
	var route *web.Route
	if route, err = interfaceToRoute(i); err != nil {
		return
	}

	path := absPath(route)
	ret, err = pathMatchPrefix(m.prefix, path)
	return
}

func (m *PrefixMatcher) Or(matchers ...matcher.Matcher) matcher.ChainableMatcher {
	return matcher.Or(m, matchers...)
}

func (m *PrefixMatcher) And(matchers ...matcher.Matcher) matcher.ChainableMatcher {
	return matcher.And(m, matchers...)
}

// PatternMatcher checks web.Route's path with given regular expresion
type RegexMatcher struct {
	regex string
}

func (m *RegexMatcher) Matches(i interface{}) (ret bool, err error) {
	var route *web.Route
	if route, err = interfaceToRoute(i); err != nil {
		return
	}

	path := absPath(route)
	ret, err = pathMatchRegex(m.regex, path)
	return
}

func (m *RegexMatcher) Or(matchers ...matcher.Matcher) matcher.ChainableMatcher {
	return matcher.Or(m, matchers...)
}

func (m *RegexMatcher) And(matchers ...matcher.Matcher) matcher.ChainableMatcher {
	return matcher.And(m, matchers...)
}
/**************************
	Constructors
***************************/
func WithMethods(methods...string) web.RouteMatcher {
	if len(methods) == 0 {
		return matcher.Any()
	}

	var mMatcher matcher.ChainableMatcher = &MethodMatcher{methods[0]}
	for _,m := range methods[1:] {
		mMatcher = mMatcher.Or(&MethodMatcher{m})
	}
	return mMatcher
}

func WithPattern(pattern string, methods...string) web.RouteMatcher {
	pMatcher := &PatternMatcher{pattern}
	mMatcher := WithMethods(methods...)
	return pMatcher.And(mMatcher)
}

func WithPrefix(prefix string, methods...string) web.RouteMatcher {
	pMatcher := &PrefixMatcher{prefix}
	mMatcher := WithMethods(methods...)
	return pMatcher.And(mMatcher)
}

func WithRegex(regex string, methods...string) web.RouteMatcher {
	pMatcher := &RegexMatcher{regex}
	mMatcher := WithMethods(methods...)
	return pMatcher.And(mMatcher)
}

/**************************
	helpers
***************************/
func interfaceToRoute(i interface{}) (*web.Route, error) {
	switch i.(type) {
	case web.Route:
		r := i.(web.Route)
		return &r, nil
	case *web.Route:
		return i.(*web.Route), nil
	default:
		return nil, fmt.Errorf("RouteMatcher doesn't support %T", i)
	}
}

func absPath(r *web.Route) string {
	p := pathutils.Join(r.Group, r.Path)
	if !pathutils.IsAbs(p) {
		p = "/" + p
	}
	return pathutils.Clean(p)
}

func methodMatch(expected, actual string) bool {
	return expected == "" || strings.ToUpper(expected) == strings.ToUpper(actual)
}

func pathMatchPattern(pattern, path string) (bool, error) {
	if pattern == "" {
		return true, nil
	}
	return doublestar.Match(pattern, path)
}

func pathMatchPrefix(prefix, path string) (bool, error) {
	return strings.HasPrefix(path, prefix), nil
}

func pathMatchRegex(regex, path string) (bool, error) {
	return regexp.MatchString(regex, path)
}