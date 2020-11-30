package matcher

import (
	"context"
	"fmt"
	"github.com/bmatcuk/doublestar"
	"regexp"
	"strings"
)

// StringMatcher is a typed ChainableMatcher that accept String
type StringMatcher interface {
	ChainableMatcher
}

// stringMatcher implements ChainableMatcher, StringMatcher and only accept String
type stringMatcher struct {
	description string
	matchFunc func(context.Context, string) (bool, error)
}

func (m *stringMatcher) StringMatches(c context.Context, value string) (bool, error) {
	return m.matchFunc(c, value)
}

func (m *stringMatcher) Matches(i interface{}) (bool, error) {
	v, ok := i.(string)
	if !ok {
		return false, fmt.Errorf("StringMatcher doesn't support %T", i)
	}
	return m.StringMatches(context.TODO(), v)
}

func (m *stringMatcher) MatchesWithContext(c context.Context, i interface{}) (bool, error) {
	v, ok := i.(string)
	if !ok {
		return false, fmt.Errorf("StringMatcher doesn't support %T", i)
	}
	return m.StringMatches(c, v)
}

func (m *stringMatcher) Or(matchers ...Matcher) ChainableMatcher {
	return Or(m, matchers...)
}

func (m *stringMatcher) And(matchers ...Matcher) ChainableMatcher {
	return And(m, matchers...)
}

func (m *stringMatcher) String() string {
	return m.description
}

/**************************
	Constructors
***************************/
func WithString(expected string, caseInsensitive bool) StringMatcher {
	desc := fmt.Sprintf("matches [%s]", expected)
	if caseInsensitive {
		desc = desc + ", case insensitive"
	}
	return &stringMatcher{
		matchFunc: func(_ context.Context, value string) (bool, error) {
			return MatchString(expected, value, caseInsensitive), nil
		},
		description: desc,
	}
}

func WithPathPattern(pattern string) StringMatcher {
	return &stringMatcher{
		matchFunc: func(_ context.Context, value string) (bool, error) {
			return MatchPathPattern(pattern, value)
		},
		description: fmt.Sprintf("matches pattern [%s]", pattern),
	}
}

func WithPrefix(prefix string, caseInsensitive bool) StringMatcher {
	desc := fmt.Sprintf("start with [%s]", prefix)
	if caseInsensitive {
		desc = desc + ", case insensitive"
	}
	return &stringMatcher{
		matchFunc: func(_ context.Context, value string) (bool, error) {
			return MatchPrefix(prefix, value, caseInsensitive)
		},
		description: desc,
	}
}

func WithRegex(regex string) StringMatcher {
	return &stringMatcher{
		matchFunc: func(_ context.Context, value string) (bool, error) {
			return MatchRegex(regex, value)
		},
		description: fmt.Sprintf("matches regex [%s]", regex),
	}
}

/**************************
	helpers
***************************/
func MatchString(expected, actual string, caseInsensitive bool) bool {
	if caseInsensitive {
		return strings.ToLower(expected) == strings.ToLower(actual)
	}
	return strings.ToUpper(expected) == strings.ToUpper(actual)
}

// MatchPathPattern given string with path pattern
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
func MatchPathPattern(pattern, path string) (bool, error) {
	if pattern == "" {
		return true, nil
	}
	return doublestar.Match(pattern, path)
}

func MatchPrefix(prefix, value string, caseInsensitive bool) (bool, error) {
	if caseInsensitive {
		return strings.HasPrefix(strings.ToLower(value), strings.ToLower(prefix)), nil
	}
	return strings.HasPrefix(value, prefix), nil
}

func MatchRegex(regex, value string) (bool, error) {
	return regexp.MatchString(regex, value)
}