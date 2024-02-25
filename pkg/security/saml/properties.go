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

package samlctx

import (
	"github.com/cisco-open/go-lanai/pkg/bootstrap"
	"github.com/pkg/errors"
)

const SamlPropertiesPrefix = "security.auth.saml"

type SamlProperties struct {
	CertificateFile string `json:"certificate-file"`
	KeyFile         string `json:"key-file"`
	KeyPassword     string `json:"key-password"`
	NameIDFormat    string `json:"name-id-format"`
}

func NewSamlProperties() *SamlProperties {
	return &SamlProperties{
		//We use this property by default so that the auth request generated by the saml package will not
		//have NameIDFormat by default
		//See saml.nameIDFormat() in github.com/crewjam/saml
		NameIDFormat: "urn:oasis:names:tc:SAML:1.1:nameid-format:unspecified",
	}
}

func BindSamlProperties(ctx *bootstrap.ApplicationContext) SamlProperties {
	props := NewSamlProperties()
	if err := ctx.Config().Bind(props, SamlPropertiesPrefix); err != nil {
		panic(errors.Wrap(err, "failed to bind SamlProperties"))
	}
	return *props
}
