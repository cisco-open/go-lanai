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

package scope

import (
	"fmt"
	"github.com/cisco-open/go-lanai/pkg/bootstrap"
	securityint "github.com/cisco-open/go-lanai/pkg/integrate/security"
	"github.com/cisco-open/go-lanai/pkg/integrate/security/seclient"
	"github.com/cisco-open/go-lanai/pkg/log"
	"github.com/cisco-open/go-lanai/pkg/security/oauth2"
	"github.com/cisco-open/go-lanai/pkg/utils"
	"go.uber.org/fx"
	"time"
)

var logger = log.New("SEC.Scope")

var Module = &bootstrap.Module{
	Name:       "security scope",
	Precedence: bootstrap.SecurityIntegrationPrecedence,
	Options: []fx.Option{
		fx.Provide(provideDefaultScopeManager),
		fx.Provide(tracingProvider()),
		fx.Invoke(configureScopeManagers),
	},
}

func Use() {
	seclient.Use()
	bootstrap.Register(Module)
}

// FxManagerCustomizer takes providers of ManagerCustomizer and wrap them with FxGroup
func FxManagerCustomizer(constructor interface{}) fx.Annotated {
	return fx.Annotated{
		Group:  FxGroup,
		Target: constructor,
	}
}

type defaultDI struct {
	fx.In
	AuthClient       seclient.AuthenticationClient             `optional:"true"`
	Properties       securityint.SecurityIntegrationProperties `optional:"true"`
	TokenStoreReader oauth2.TokenStoreReader                   `optional:"true"`
	Customizers      []ManagerCustomizer                       `group:"security-scope"`
}

func provideDefaultScopeManager(di defaultDI) (ScopeManager, error) {
	if di.TokenStoreReader == nil || di.AuthClient == nil {
		return nil, fmt.Errorf(`security scope managers requires "resserver" and "seclient", but not configured`)
	}

	// default options
	opts := []ManagerOptions{
		func(opt *managerOption) {
			opt.Client = di.AuthClient
			opt.TokenStoreReader = di.TokenStoreReader
			opt.BackOffPeriod = time.Duration(di.Properties.FailureBackOff)
			opt.GuaranteedValidity = time.Duration(di.Properties.GuaranteedValidity)

			// parse accounts
			credentials := map[string]string{}
			sysAccts := utils.NewStringSet()
			if di.Properties.Accounts.Default.Username != "" {
				opt.DefaultSystemAccount = di.Properties.Accounts.Default.Username
				credentials[di.Properties.Accounts.Default.Username] = di.Properties.Accounts.Default.Password
				sysAccts.Add(di.Properties.Accounts.Default.Username)
			}
			// TBD, this is consistent behavior from java impl. Such configuration allows dev-ops to give
			// special treatment on certain accounts. Since we don't know any use case of this feature at
			// the time of writing this code, we temporarily disabled it, but keep the code for reference.
			//for _, acct := range di.Properties.Accounts.Additional {
			//	if acct.UName == "" || acct.Password == "" {
			//		continue
			//	}
			//	credentials[acct.UName] = acct.Password
			//	if acct.SystemAccount {
			//		sysAccts.Add(acct.UName)
			//	}
			//}
			opt.KnownCredentials = credentials
			opt.SystemAccounts = sysAccts
		},
	}

	// customizers
	for _, c := range di.Customizers {
		opts = append(opts, c.Customize()...)
	}

	return newDefaultScopeManager(opts...), nil
}

func configureScopeManagers(EffectiveScopeManager ScopeManager) {
	scopeManager = EffectiveScopeManager
}
