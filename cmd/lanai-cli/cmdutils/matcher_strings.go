// Copyright 2023 Cisco Systems, Inc. and its affiliates
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package cmdutils

import (
    "context"
    "fmt"
    "github.com/bmatcuk/doublestar/v4"
    "regexp"
    "strings"
)

const (
    descSuffixCaseInsensitive = `, case insensitive`
)

// StringMatcher is a typed ChainableMatcher that accept String
type StringMatcher interface {
    ChainableMatcher[string]
}

/**************************
	Constructors
***************************/

func WithString(expected string, caseInsensitive bool) StringMatcher {
    desc := fmt.Sprintf("matches [%s]", expected)
    if caseInsensitive {
        desc = desc + descSuffixCaseInsensitive
    }
    return &GenericMatcher[string]{
        MatchFunc: func(_ context.Context, value string) (bool, error) {
            return MatchString(expected, value, caseInsensitive), nil
        },
        Description: desc,
    }
}

func WithSubString(substr string, caseInsensitive bool) StringMatcher {
    desc := fmt.Sprintf("contains [%s]", substr)
    if caseInsensitive {
        desc = desc + descSuffixCaseInsensitive
    }
    return &GenericMatcher[string]{
        MatchFunc: func(_ context.Context, value string) (bool, error) {
            return MatchSubString(substr, value, caseInsensitive), nil
        },
        Description: desc,
    }
}

func AnyNonEmptyString() StringMatcher {
    desc := fmt.Sprintf("matches any non-empty string")
    return &GenericMatcher[string]{
        MatchFunc: func(_ context.Context, value string) (bool, error) {
            return value != "", nil
        },
        Description: desc,
    }
}

func WithPathPattern(pattern string) StringMatcher {
    return &GenericMatcher[string]{
        MatchFunc: func(_ context.Context, value string) (bool, error) {
            return MatchPathPattern(pattern, value)
        },
        Description: fmt.Sprintf("matches pattern [%s]", pattern),
    }
}

func WithPrefix(prefix string, caseInsensitive bool) StringMatcher {
    desc := fmt.Sprintf("start with [%s]", prefix)
    if caseInsensitive {
        desc = desc + descSuffixCaseInsensitive
    }
    return &GenericMatcher[string]{
        MatchFunc: func(_ context.Context, value string) (bool, error) {
            return MatchPrefix(prefix, value, caseInsensitive)
        },
        Description: desc,
    }
}

func WithSuffix(suffix string, caseInsensitive bool) StringMatcher {
    desc := fmt.Sprintf("ends with [%s]", suffix)
    if caseInsensitive {
        desc = desc + descSuffixCaseInsensitive
    }
    return &GenericMatcher[string]{
        MatchFunc: func(_ context.Context, value string) (bool, error) {
            return MatchSuffix(suffix, value, caseInsensitive)
        },
        Description: desc,
    }
}

func WithRegex(regex string) StringMatcher {
    return &GenericMatcher[string]{
        MatchFunc: func(_ context.Context, value string) (bool, error) {
            return MatchRegex(regex, value)
        },
        Description: fmt.Sprintf("matches regex [%s]", regex),
    }
}

func WithRegexPattern(regex *regexp.Regexp) StringMatcher {
    return &GenericMatcher[string]{
        MatchFunc: func(_ context.Context, value string) (bool, error) {
            return MatchRegexPattern(regex, value)
        },
        Description: fmt.Sprintf("matches regex [%s]", regex.String()),
    }
}

/**************************
	helpers
***************************/

func MatchString(expected, actual string, caseInsensitive bool) bool {
    if caseInsensitive {
        expected = strings.ToLower(expected)
        actual = strings.ToLower(actual)
    }
    return expected == actual
}

func MatchSubString(substr, actual string, caseInsensitive bool) bool {
    if caseInsensitive {
        substr = strings.ToLower(substr)
        actual = strings.ToLower(actual)
    }
    return strings.Contains(actual, substr)
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

func MatchSuffix(suffix, value string, caseInsensitive bool) (bool, error) {
    if caseInsensitive {
        return strings.HasSuffix(strings.ToLower(value), strings.ToLower(suffix)), nil
    }
    return strings.HasSuffix(value, suffix), nil
}

func MatchRegex(regex, value string) (bool, error) {
    return regexp.MatchString(regex, value)
}

func MatchRegexPattern(regex *regexp.Regexp, value string) (bool, error) {
    return regex.MatchString(value), nil
}
