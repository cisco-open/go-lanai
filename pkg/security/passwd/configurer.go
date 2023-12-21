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

package passwd

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/redis"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"fmt"
)

var (
	PasswordAuthenticatorFeatureId = security.FeatureId("passwdAuth", security.FeatureOrderAuthenticator)
)

type PasswordAuthConfigurer struct {
	accountStore security.AccountStore
	passwordEncoder PasswordEncoder
	redisClient redis.Client
}

func newPasswordAuthConfigurer(store security.AccountStore, encoder PasswordEncoder, redisClient redis.Client) *PasswordAuthConfigurer {
	return &PasswordAuthConfigurer {
		accountStore:    store,
		passwordEncoder: encoder,
		redisClient:     redisClient,
	}
}

func (pac *PasswordAuthConfigurer) Apply(feature security.Feature, ws security.WebSecurity) error {
	// Verify
	if err := pac.validate(feature.(*PasswordAuthFeature), ws); err != nil {
		return err
	}
	f := feature.(*PasswordAuthFeature)

	// Build authenticator
	ctx := context.Background()
	defaults := &builderDefaults{
		accountStore: pac.accountStore,
		passwordEncoder: pac.passwordEncoder,
		redisClient: pac.redisClient,
	}
	authenticator, err := NewAuthenticatorBuilder(f, defaults).Build(ctx)
	if err != nil {
		return err
	}

	// Add authenticator to WS, flatten if multiple
	if composite, ok := authenticator.(*security.CompositeAuthenticator); ok {
		ws.Authenticator().(*security.CompositeAuthenticator).Merge(composite)
	} else {
		ws.Authenticator().(*security.CompositeAuthenticator).Add(authenticator)
	}

	return nil
}

func (pac *PasswordAuthConfigurer) validate(f *PasswordAuthFeature, ws security.WebSecurity) error {

	if _,ok := ws.Authenticator().(*security.CompositeAuthenticator); !ok {
		return fmt.Errorf("unable to add password authenticator to %T", ws.Authenticator())
	}

	if f.accountStore == nil && pac.accountStore == nil {
		return fmt.Errorf("unable to create password authenticator: account accountStore is not set")
	}
	return nil
}



