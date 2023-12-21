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

package authserver

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"github.com/pkg/errors"
)

const (
	PropertiesPrefix = "security.auth"
)

//goland:noinspection GoNameStartsWithPackageName
type AuthServerProperties struct {
	Issuer            IssuerProperties    `json:"issuer"`
	RedirectWhitelist []string            `json:"redirect-whitelist"`
	Endpoints         EndpointsProperties `json:"endpoints"`
}

type IssuerProperties struct {
	//  the protocol which is either http or https
	Protocol string `json:"protocol"`
	// This server's host name
	// Used to build the entity base url. The entity url identifies this auth server in a SAML exchange and OIDC exchange.
	Domain string `json:"domain"`
	Port   int    `json:"port"`
	// Context base path for this server
	// Used to build the entity base url. The entity url identifies this auth server in a SAML exchange.
	ContextPath string `json:"context-path"`
	IncludePort bool   `json:"include-port"`
}

type EndpointsProperties struct {
	// TODO check_session is necessary and should be implemented. Java: SessionInfoEndpoint
	Authorize       string `json:"authorize"`
	Token           string `json:"token"`
	Approval        string `json:"approval"`
	CheckToken      string `json:"check-token"`
	TenantHierarchy string `json:"tenant-hierarchy"`
	Error           string `json:"error"`
	Logout          string `json:"logout"`
	LoggedOut       string `json:"logged-out"`
	UserInfo        string `json:"user-info"`
	JwkSet          string `json:"jwk-set"`
	SamlMetadata    string `json:"saml-metadata"`
}

//NewAuthServerProperties create a SessionProperties with default values
func NewAuthServerProperties() *AuthServerProperties {
	return &AuthServerProperties{
		Issuer: IssuerProperties{
			Protocol:    "http",
			Domain:      "locahost",
			Port:        8080,
			ContextPath: "",
			IncludePort: true,
		},
		RedirectWhitelist: []string{},
		Endpoints: EndpointsProperties{
			Authorize:       "/v2/authorize",
			Token:           "/v2/token",
			Approval:        "/v2/approve",
			CheckToken:      "/v2/check_token",
			TenantHierarchy: "/v2/tenant_hierarchy",
			Error:           "/error",
			Logout:          "/v2/logout",
			UserInfo:        "/v2/userinfo",
			JwkSet:          "/v2/jwks",
			SamlMetadata:    "/metadata",
			LoggedOut:       "/",
		},
	}
}

//BindAuthServerProperties create and bind AuthServerProperties, with a optional prefix
func BindAuthServerProperties(ctx *bootstrap.ApplicationContext) AuthServerProperties {
	props := NewAuthServerProperties()
	if err := ctx.Config().Bind(props, PropertiesPrefix); err != nil {
		panic(errors.Wrap(err, "failed to bind AuthServerProperties"))
	}
	return *props
}
