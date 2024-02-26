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
	"github.com/cisco-open/go-lanai/pkg/bootstrap"
	"github.com/cisco-open/go-lanai/pkg/log"
	"github.com/pkg/errors"
	"strings"
)

var logger = log.New("OAuth2.JWT")

/***********************
	Crypto
************************/

const CryptoKeysPropertiesPrefix = "security"

const (
	KeyFileFormatPem KeyFormatType = "pem"
)

type KeyFormatType string

type CryptoProperties struct {
	Keys map[string]CryptoKeyProperties `json:"keys"`
	Jwt JwtProperties `json:"jwt"`
}

type JwtProperties struct {
	KeyName string `json:"key-name"`
}

type CryptoKeyProperties struct {
	Id        string `json:"id"`
	KeyFormat string `json:"format"`
	Location  string `json:"file"`
	Password  string `json:"password"`
}

func (p CryptoKeyProperties) Format() KeyFormatType {
	return KeyFormatType(strings.ToLower(p.KeyFormat))
}

//CryptoProperties create a SessionProperties with default values
func NewCryptoProperties() *CryptoProperties {
	return &CryptoProperties {
		Keys: map[string]CryptoKeyProperties{},
	}
}

//BindCryptoProperties create and bind CryptoProperties, with a optional prefix
func BindCryptoProperties(ctx *bootstrap.ApplicationContext) CryptoProperties {
	props := NewCryptoProperties()
	if err := ctx.Config().Bind(props, CryptoKeysPropertiesPrefix); err != nil {
		panic(errors.Wrap(err, "failed to bind CryptoProperties"))
	}
	return *props
}
