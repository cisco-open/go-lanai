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

package service

import (
	"context"
	samlctx "cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/saml"
	samlidp "cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/saml/idp"
	"errors"
)

type InMemorySamlClientStore struct {
	details []samlidp.DefaultSamlClient
}

func NewInMemSpManager() samlctx.SamlClientStore {
	return &InMemorySamlClientStore{
		details: []samlidp.DefaultSamlClient{
			//TODO: populate your SAML SP details here
		},
	}
}

func (i *InMemorySamlClientStore) GetAllSamlClient(context.Context) ([]samlctx.SamlClient, error) {
	var result []samlctx.SamlClient
	for _, v := range i.details {
		result = append(result, v)
	}
	return result, nil
}

func (i *InMemorySamlClientStore) GetSamlClientByEntityId(ctx context.Context, entityId string) (samlctx.SamlClient, error) {
	for _, detail := range i.details {
		if detail.EntityId == entityId {
			return detail, nil
		}
	}
	return samlidp.DefaultSamlClient{}, errors.New("not found")
}
