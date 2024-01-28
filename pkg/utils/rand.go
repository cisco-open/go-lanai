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

package utils

import (
	cryptorand "crypto/rand"
	"math/big"
	"math/rand"
)

const (
	CharsetAlphanumeric RandomCharset = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	CharsetAlphabetic   RandomCharset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

func init() {
	//rand.Seed(time.Now().UnixNano())
}

// RandomCharset is a string containing all acceptable UTF-8 characters for random string generation
type RandomCharset string

// RandomString returns a random Alphanumeric string of given "length"
// this function uses "crypto/rand" and fallback to "math/rand"
// It panics if len(charset) > 255, and returns empty string if length is non-positive
func RandomString(length int) string {
	return RandomStringWithCharset(length, CharsetAlphanumeric)
}

// RandomStringWithCharset returns a random string of given "length" containing only characters from given "charset"
// this function uses "crypto/rand" and fallback to "math/rand"
// It returns empty string if length is non-positive, and only the first 256 chars in "charset" are used
func RandomStringWithCharset(length int, charset RandomCharset) string {
	if length <= 0 {
		return ""
	}

	data := make([]byte, length)
	b := make([]byte, 1)
	for i := range data {
		if n, e := cryptorand.Reader.Read(b); e != nil || n < 1 {
			data[i] = charset[rand.Intn(len(charset))] //nolint:gosec // this is fallback method, better than not working
		} else {
			data[i] = charset[int(b[0]) % len(charset)]
		}
	}
	return string(data)
}

// RandomInt64N returns, as an int64, a non-negative uniform number in the half-open interval [0,n).
// This function uses "crypto/rand" and fallback to "math/rand".
// It panics if n <= 0.
func RandomInt64N(n int64) int64 {
	bigInt, e := cryptorand.Int(cryptorand.Reader, big.NewInt(n))
	if e != nil {
		return rand.Int63n(n) //nolint:gosec // this is fallback method, better than not working
	}
	return bigInt.Int64()
}

// RandomIntN returns, as an int64, a non-negative uniform number in the half-open interval [0,n).
// This function uses "crypto/rand" and fallback to "math/rand".
// It panics if n <= 0.
func RandomIntN(n int) int {
	return int(RandomInt64N(int64(n)))
}
