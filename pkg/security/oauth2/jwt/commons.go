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

package jwt

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"github.com/golang-jwt/jwt/v4"
	"strings"
)

// ParseJwtHeaders extract JWT's headers without verifying the token
func ParseJwtHeaders(jwtValue string) (map[string]interface{}, error) {
	parts := strings.Split(jwtValue, ".")
	if len(parts) != 3 {
		return nil, jwt.NewValidationError("token contains an invalid number of segments", jwt.ValidationErrorMalformed)
	}

	// parse Header
	//nolint:staticcheck
	headerBytes, e := jwt.DecodeSegment(parts[0])
	if e != nil {
		return nil, &jwt.ValidationError{Inner: e, Errors: jwt.ValidationErrorMalformed}
	}

	var headers map[string]interface{}
	if e := json.Unmarshal(headerBytes, &headers); e != nil {
		return nil, &jwt.ValidationError{Inner: e, Errors: jwt.ValidationErrorMalformed}
	}
	return headers, nil
}

func printPrivateKey(key *rsa.PrivateKey) string {
	//bytes := x509.MarshalPKCS1PrivateKey(key)
	bytes, _ := x509.MarshalPKCS8PrivateKey(key)
	return base64.StdEncoding.EncodeToString(bytes)
}

func printPublicKey(key *rsa.PublicKey) string {
	bytes := x509.MarshalPKCS1PublicKey(key)
	return base64.StdEncoding.EncodeToString(bytes)
}
