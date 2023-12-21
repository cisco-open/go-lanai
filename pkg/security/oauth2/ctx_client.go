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

package oauth2

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"time"
)

/***********************************
	DTO
 ***********************************/

//goland:noinspection GoNameStartsWithPackageName
type OAuth2Client interface {
	ClientId() string
	SecretRequired() bool
	Secret() string
	GrantTypes() utils.StringSet
	RedirectUris() utils.StringSet
	Scopes() utils.StringSet
	AutoApproveScopes() utils.StringSet
	AccessTokenValidity() time.Duration
	RefreshTokenValidity() time.Duration
	UseSessionTimeout() bool
	AssignedTenantIds() utils.StringSet
	ResourceIDs() utils.StringSet
}

/***********************************
	Store
 ***********************************/

//goland:noinspection GoNameStartsWithPackageName
type OAuth2ClientStore interface {
	LoadClientByClientId(ctx context.Context, clientId string) (OAuth2Client, error)
}
