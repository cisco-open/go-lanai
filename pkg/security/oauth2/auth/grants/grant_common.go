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

package grants

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/auth"
)

var logger = log.New("OAuth2.Grant")

func CommonPreGrantValidation(c context.Context, client oauth2.OAuth2Client, request *auth.TokenRequest) error {
	// check grant
	if e := auth.ValidateGrant(c, client, request.GrantType); e != nil {
		return e
	}

	// check scope
	if e := auth.ValidateAllScopes(c, client, request.Scopes); e != nil {
		return e
	}
	return nil
}

