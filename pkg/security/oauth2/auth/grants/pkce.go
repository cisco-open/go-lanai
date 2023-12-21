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

package grants

import (
	"crypto"
	"encoding/base64"
	"fmt"
	"strings"
)

const (
	PKCEChallengeMethodPlain  PKCECodeChallengeMethod = "plain"
	PKCEChallengeMethodSHA256 PKCECodeChallengeMethod = "S256"
)

type PKCECodeChallengeMethod string

func (m *PKCECodeChallengeMethod) UnmarshalText(text []byte) error {
	str := string(text)
	switch {
	case string(PKCEChallengeMethodPlain) == strings.ToLower(str):
		*m = PKCEChallengeMethodPlain
	case string(PKCEChallengeMethodSHA256) == strings.ToUpper(str):
		*m = PKCEChallengeMethodSHA256
	case len(text) == 0:
		*m = PKCEChallengeMethodPlain
	default:
		return fmt.Errorf("invalid code challenge method")
	}
	return nil
}

func parseCodeChallengeMethod(str string) (ret PKCECodeChallengeMethod, err error) {
	err = ret.UnmarshalText([]byte(str))
	return
}

// https://datatracker.ietf.org/doc/html/rfc7636#section-4.6
func verifyPKCE(toVerify string, challenge string, method PKCECodeChallengeMethod) (ret bool) {
	var encoded string
	switch method {
	case PKCEChallengeMethodPlain:
		encoded = toVerify
	case PKCEChallengeMethodSHA256:
		hash := crypto.SHA256.New()
		if _, e := hash.Write([]byte(toVerify)); e != nil {
			return
		}
		encoded = base64.RawURLEncoding.EncodeToString(hash.Sum(nil))
	default:
		return
	}
	return encoded == challenge
}