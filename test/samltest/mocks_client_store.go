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

import (
    "context"
    "errors"
    "github.com/cisco-open/go-lanai/pkg/bootstrap"
    samlctx "github.com/cisco-open/go-lanai/pkg/security/saml"
    "github.com/crewjam/saml"
)

type ClientStoreMockOptions func(opt *ClientStoreMockOption)
type ClientStoreMockOption struct {
	Clients []samlctx.SamlClient
	SPs []*saml.ServiceProvider
	ClientsProperties map[string]MockedClientProperties
}

// ClientsWithPropertiesPrefix returns a ClientStoreMockOptions that bind a map of properties from application config with given prefix
func ClientsWithPropertiesPrefix(appCfg bootstrap.ApplicationConfig, prefix string) ClientStoreMockOptions {
	return func(opt *ClientStoreMockOption) {
		if e := appCfg.Bind(&opt.ClientsProperties, prefix); e != nil {
			panic(e)
		}
	}
}

// ClientsWithSPs returns a ClientStoreMockOptions that convert given SPs to Clients
func ClientsWithSPs(sps...*saml.ServiceProvider) ClientStoreMockOptions {
	return func(opt *ClientStoreMockOption) {
		opt.SPs = sps
	}
}

type MockSamlClientStore struct {
	details []samlctx.SamlClient
}

func NewMockedClientStore(opts...ClientStoreMockOptions) *MockSamlClientStore {
	opt := ClientStoreMockOption {}
	for _, fn := range opts {
		fn(&opt)
	}

	var details []samlctx.SamlClient
	switch {
	case len(opt.Clients) > 0:
		details = opt.Clients
	case len(opt.SPs) > 0:
		for _, sp := range opt.SPs {
			v := NewMockedSamlClient(func(opt *MockedClientOption) {
				opt.SP = sp
			})
			details = append(details, v)
		}
	default:
		for _, props := range opt.ClientsProperties {
			v := NewMockedSamlClient(func(opt *MockedClientOption) {
				opt.Properties = props
			})
			details = append(details, v)
		}
	}

	return &MockSamlClientStore{details: details}
}

func (t *MockSamlClientStore) GetAllSamlClient(_ context.Context) ([]samlctx.SamlClient, error) {
	var result []samlctx.SamlClient
	for _, v := range t.details {
		result = append(result, v)
	}
	return result, nil
}

func (t *MockSamlClientStore) GetSamlClientByEntityId(_ context.Context, id string) (samlctx.SamlClient, error) {
	for _, detail := range t.details {
		if detail.GetEntityId() == id {
			return detail, nil
		}
	}
	return nil, errors.New("not found")
}

