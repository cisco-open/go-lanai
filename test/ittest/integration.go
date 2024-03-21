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

package ittest

import (
	"github.com/cisco-open/go-lanai/pkg/integrate/httpclient"
	secit "github.com/cisco-open/go-lanai/pkg/integrate/security"
	"github.com/cisco-open/go-lanai/pkg/integrate/security/scope"
	"github.com/cisco-open/go-lanai/pkg/integrate/security/seclient"
	"github.com/cisco-open/go-lanai/pkg/security/oauth2"
	"github.com/cisco-open/go-lanai/test"
	"github.com/cisco-open/go-lanai/test/apptest"
	"go.uber.org/fx"
	"gopkg.in/dnaeon/go-vcr.v3/cassette"
	"gopkg.in/dnaeon/go-vcr.v3/recorder"
	"net/http"
	"strings"
	"time"
)

func WithRecordedScopes() test.Options {
	fxOpts := []fx.Option{
		fx.Provide(provideScopeDI),
		fx.Provide(provideScopeVCROptions),
	}

	opts := []test.Options{
		apptest.WithModules(scope.Module, seclient.Module),
		apptest.WithFxOptions(fxOpts...),
	}
	return func(opt *test.T) {
		for _, fn := range opts {
			fn(opt)
		}
	}
}

/*************************
	Providers
 *************************/

type scopeDI struct {
	fx.In
	ItProperties secit.SecurityIntegrationProperties
	HttpClient   httpclient.Client
	Recorder     *recorder.Recorder `optional:"true"`
}

type scopeDIOut struct {
	fx.Out
	TokenReader oauth2.TokenStoreReader
}

func provideScopeDI(di scopeDI) scopeDIOut {
	tokenReader := NewRemoteTokenStoreReader(func(opt *RemoteTokenStoreOption) {
		opt.SkipRemoteCheck = true
		opt.HttpClient = di.HttpClient
		opt.BaseUrl = di.ItProperties.Endpoints.BaseUrl
		opt.ServiceName = di.ItProperties.Endpoints.ServiceName
		opt.Scheme = di.ItProperties.Endpoints.Scheme
		opt.ContextPath = di.ItProperties.Endpoints.ContextPath
		opt.ClientId = di.ItProperties.Client.ClientId
		opt.ClientSecret = di.ItProperties.Client.ClientSecret
		if di.Recorder != nil {
			opt.HttpClientConfig = &httpclient.ClientConfig{
				HTTPClient: di.Recorder.GetDefaultClient(),
			}
		}
	})
	return scopeDIOut{
		TokenReader: tokenReader,
	}
}

type scopeVCROptionsOut struct {
	fx.Out
	VCROptions HTTPVCROptions `group:"http-vcr"`
}

func provideScopeVCROptions() scopeVCROptionsOut {
	return scopeVCROptionsOut{
		VCROptions: HttpRecorderHooks(NewRecorderHook(extendedTokenValidityHook(), recorder.BeforeResponseReplayHook)),
	}
}

/*************************
	Additional Hooks
 *************************/

// extendedTokenValidityHook HTTP VCR hook that extend token validity to a distant future.
// During scope switching, token's expiry time is used to determine if token need to be refreshed.
// This would cause inconsistent HTTP interactions between recording time and replay time (after token expires)
// "expiry" and "expires_in" are JSON fields in `/v2/token` response and `exp` is a standard claim in `/v2/check_token` response
func extendedTokenValidityHook() func(i *cassette.Interaction) error {
	longValidity := 100 * 24 * 365 * time.Hour
	expiry := time.Now().Add(longValidity)
	tokenBodySanitizers := map[string]ValueSanitizer{
		"expiry":     SubstituteValueSanitizer(expiry.Format(time.RFC3339)),
		"expires_in": SubstituteValueSanitizer(longValidity.Seconds()),
		"exp":        SubstituteValueSanitizer(expiry.Unix()),
	}
	tokenBodyJsonPaths := parseJsonPaths([]string{"$.expiry", "$.expires_in", "$.exp"})
	return func(i *cassette.Interaction) error {
		if i.Response.Code != http.StatusOK ||
			!strings.Contains(i.Request.URL, "/v2/token") && !strings.Contains(i.Request.URL, "/v2/check_token") {
			return nil
		}
		i.Response.Body = sanitizeJsonBody(i.Response.Body, tokenBodySanitizers, tokenBodyJsonPaths)
		return nil
	}
}
