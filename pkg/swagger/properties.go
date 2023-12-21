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

package swagger

const SwaggerPrefix = "swagger"

type SwaggerProperties struct {
	BasePath string                    `json:"base-path"`
	Spec     string                    `json:"spec"`
	Security SwaggerSecurityProperties `json:"security"`
	UI       SwaggerUIProperties       `json:"ui"`
}

type SwaggerSecurityProperties struct {
	SecureDocs bool                 `json:"secure-docs"`
	Sso        SwaggerSsoProperties `json:"sso"`
}

type SwaggerSsoProperties struct {
	BaseUrl          string                `json:"base-url"`
	TokenPath        string                `json:"token-path"`
	AuthorizePath    string                `json:"authorize-path"`
	ClientId         string                `json:"client-id"`
	ClientSecret     string                `json:"client-secret"`
	AdditionalParams []ParameterProperties `json:"additional-params" binding:"omitempty"`
}

type ParameterProperties struct {
	Name               string `json:"name"`
	DisplayName        string `json:"display-name"`
	CandidateSourceUrl string `json:"candidate-source-url"`
	CandidateJsonPath  string `json:"candidate-json-path"`
}

type SwaggerUIProperties struct {
	Title string `json:"title"`
}

func NewSwaggerSsoProperties() *SwaggerProperties {
	return &SwaggerProperties{}
}
