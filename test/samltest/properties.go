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

package samltest

import "cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"

type ProviderProperties struct {
	EntityID         string `json:"entity-id"`
	MetadataSource   string `json:"metadata-source"`
	CertsSource      string `json:"certs"`
	PrivateKeySource string `json:"private-key"`
}

type ExtIDPProperties struct {
	Domain string `json:"domain"`
	Name   string `json:"name"`
	IdName string `json:"id-name"`
}

type IDPProperties struct {
	ProviderProperties
	ExtIDPProperties
	SSOPath string `json:"sso"`
	SLOPath string `json:"slo"`
}

type SPProperties struct {
	ProviderProperties
	ACSPath string         `json:"acs"`
	SLOPath string         `json:"slo"`
	IDP     *IDPProperties `json:"idp"`
}

type MockedClientProperties struct {
	SPProperties
	SkipEncryption            bool                      `json:"skip-encryption"`
	SkipSignatureVerification bool                      `json:"skip-signature-verification"`
	TenantRestriction         utils.CommaSeparatedSlice `json:"tenant-restriction"`
	TenantRestrictionType     string                    `json:"tenant-restriction-type"`
}
