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

package passwd

import "golang.org/x/crypto/bcrypt"

type PasswordEncoder interface {
	Encode(rawPassword string) string
	Matches(raw, encoded string) bool
}


type noopPasswordEncoder string

func NewNoopPasswordEncoder() PasswordEncoder {
	return noopPasswordEncoder("clear text")
}
func (noopPasswordEncoder) Encode(rawPassword string) string {
	return rawPassword
}

func (noopPasswordEncoder) Matches(raw, encoded string) bool {
	return raw == encoded
}

// bcryptPasswordEncoder implements PasswordEncoder
type bcryptPasswordEncoder struct {
	cost int
}

func NewBcryptPasswordEncoder() PasswordEncoder {
	return &bcryptPasswordEncoder{
		cost: 10,
	}
}

func (enc *bcryptPasswordEncoder) Encode(raw string) string {
	encoded, e := bcrypt.GenerateFromPassword([]byte(raw), enc.cost)
	if e != nil {
		return ""
	}
	return string(encoded)
}

func (enc *bcryptPasswordEncoder) Matches(raw, encoded string) bool {
	e := bcrypt.CompareHashAndPassword([]byte(encoded), []byte(raw))
	return e == nil
}