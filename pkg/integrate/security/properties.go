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

package security

import (
	"embed"
	"github.com/cisco-open/go-lanai/pkg/bootstrap"
	"github.com/cisco-open/go-lanai/pkg/utils"
	"github.com/pkg/errors"
	"time"
)

const (
	PropertiesPrefix = "integrate.security"
)

//go:embed defaults-integrate-security.yml
var DefaultConfigFS embed.FS

//goland:noinspection GoNameStartsWithPackageName
type SecurityIntegrationProperties struct {
	// How much time after a failed attempt, when re-try is allowed. Before this period pass,
	// integration framework will not re-attempt switching context to same combination of username and tenant name
	FailureBackOff utils.Duration `json:"failure-back-off"`

	// How much time that security context is guaranteed to be valid after requested.
	// when such validity cannot be guaranteed (e.g. this value is longer than token's validity),
	// we use FailureBackOff and re-request new token after `back-off` passes
	GuaranteedValidity utils.Duration `json:"guaranteed-validity"`

	Endpoints AuthEndpointsProperties     `json:"endpoints"`
	Client    ClientCredentialsProperties `json:"client"`
	Accounts  AccountsProperties          `json:"accounts"`
}

type ClientCredentialsProperties struct {
	ClientId     string `json:"client-id"`
	ClientSecret string `json:"secret"`
}

type AuthEndpointsProperties struct {
	// BaseUrl is used to override service discovery and load-balancing
	// When set, ServiceName, Scheme and ContextPath are ignored
	BaseUrl string `json:"base-url"`
	// ServiceName The name of auth service, used by service discovery to authentication/authorization URL
	ServiceName string `json:"service-name"`
	// Scheme HTTP scheme ("http"/"https") of auth service, in case it's not resolvable from service registry
	Scheme string `json:"scheme"`
	// ContextPath The path prefix of all endpoints, in case it's not resolvable from service registry
	ContextPath string `json:"context-path"`
	// PasswordLogin Path of password login endpoint
	PasswordLogin string `json:"password-login"`
	// SwitchContext Path of switch tenant/user endpoint
	SwitchContext string `json:"switch-context"`
}

type AccountsProperties struct {
	Default    AccountCredentialsProperties   `json:"default"`
	Additional []AccountCredentialsProperties `json:"additional"`
}

type AccountCredentialsProperties struct {
	Username      string `json:"username"`
	Password      string `json:"password"`
	SystemAccount bool   `json:"system-account"`
}

// NewSecurityIntegrationProperties create a DataProperties with default values
func NewSecurityIntegrationProperties() *SecurityIntegrationProperties {
	return &SecurityIntegrationProperties{
		FailureBackOff:     utils.Duration(300 * time.Second),
		GuaranteedValidity: utils.Duration(30 * time.Second),
		Endpoints: AuthEndpointsProperties{
			ServiceName:   "authservice",
			Scheme:        "http",
			PasswordLogin: "/v2/token",
			SwitchContext: "/v2/token",
		},
		Client: ClientCredentialsProperties{
			ClientId:     "nfv-service",
			ClientSecret: "nfv-service-secret",
		},
		Accounts: AccountsProperties{
			Default: AccountCredentialsProperties{
				Username:      "system",
				Password:      "system",
				SystemAccount: true,
			},
		},
	}
}

// BindSecurityIntegrationProperties create and bind SessionProperties, with a optional prefix
func BindSecurityIntegrationProperties(ctx *bootstrap.ApplicationContext) SecurityIntegrationProperties {
	props := NewSecurityIntegrationProperties()
	if err := ctx.Config().Bind(props, PropertiesPrefix); err != nil {
		panic(errors.Wrap(err, "failed to bind SecurityIntegrationProperties"))
	}
	return *props
}
