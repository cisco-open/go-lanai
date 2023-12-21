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

package testdata

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/idp"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/samltest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/sectest"
)

var DefaultIdpProviders = []idp.IdentityProvider {
	samltest.MockedIdpProvider {
		ExtSamlMetadata: samltest.ExtSamlMetadata{
			EntityId:         "http://www.okta.com/exkwj65c2kC1vwtYi0h7",
			Domain:           "saml.vms.com",
			Source:           "testdata/okta_login_test_metadata.xml",
			Name:             "okta",
			IdName:           "email",
		},
	},
	samltest.MockedIdpProvider{
		ExtSamlMetadata: samltest.ExtSamlMetadata{
			EntityId:         "http://www.okta.com/exk668ha29xaI4in25d7",
			Domain:           "saml-alt.vms.com",
			Source:           "testdata/okta_logout_test_metadata.xml",
			Name:             "okta",
			IdName:           "email",
		},
	},
}

var DefaultFedUserProperties = []*sectest.MockedFederatedUserProperties {
	{
		ExtIdpName:              "okta",
		ExtIdName:               "email",
		ExtIdValue:              "test@example.com",
	},
}