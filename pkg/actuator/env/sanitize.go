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

package env

import (
	"context"
	"github.com/cisco-open/go-lanai/pkg/utils"
	"github.com/cisco-open/go-lanai/pkg/utils/matcher"
	"regexp"
	"strings"
)

const (
	regexChars = "*$^+"
)

var (
	DefaultKeysToSanitize = utils.NewStringSet(
		`.*password.*`, `.*secret.*`,
		`key`, `.*credentials.*`,
		`vcap_services`, `sun.java.command`,
	)
)

type Sanitizer struct {
	keyMatcher matcher.StringMatcher
}

func NewSanitizer(keyPatterns []string) *Sanitizer {
	patterns := DefaultKeysToSanitize.Copy().Add(keyPatterns...)
	var keyMatcher matcher.StringMatcher
	for p, _ := range patterns {
		var m matcher.StringMatcher
		if isRegex(p) {
			regex := regexp.MustCompile(p)
			m = matcher.WithRegexPattern(regex)
		} else {
			m = matcher.WithString(p, false).Or(matcher.WithSuffix(p, false))
		}

		if keyMatcher == nil {
			keyMatcher = m
		} else {
			keyMatcher = keyMatcher.Or(m)
		}
	}
	return &Sanitizer{
		keyMatcher: keyMatcher,
	}
}

func (s Sanitizer) Sanitize(ctx context.Context, key string, value interface{}) interface{} {
	// 1. can we sanitize?
	switch value.(type) {
	case string, []string, utils.StringSet:
	default:
		return value
	}

	// 2. does key match?
	if ok, e := s.keyMatcher.MatchesWithContext(ctx, key); e != nil || !ok {
		return value
	}
	return "********"
}

func isRegex(s string) bool {
	return strings.ContainsAny(s, regexChars)
}