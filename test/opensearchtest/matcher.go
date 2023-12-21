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

package opensearchtest

import (
	"bytes"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/httpvcr/cassette"
	"io"
	"net/http"
)

// MatchBody will ensure that the Matcher also matches the contents of the body.
// The contents of the body can be modified by the MatcherBodyModifier before the
// comparison occurs.
func MatchBody(modifiers *MatcherBodyModifiers) cassette.Matcher {
	return func(r *http.Request, i cassette.Request) bool {
		if r.Body == nil {
			return cassette.DefaultMatcher(r, i)
		}
		var b bytes.Buffer
		if _, err := b.ReadFrom(r.Body); err != nil {
			return false
		}
		r.Body = io.NopCloser(&b)
		requestBody := b.Bytes()
		recordingBody := []byte(i.Body)
		if modifiers != nil {
			for _, modifier := range modifiers.Modifier() {
				modifier(&requestBody)
				modifier(&recordingBody)
			}
		}
		return cassette.DefaultMatcher(r, i) &&
			(string(requestBody) == "" || string(requestBody) == string(recordingBody))
	}
}
