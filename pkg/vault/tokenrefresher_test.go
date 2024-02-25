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

package vault_test

import (
	"context"
	"github.com/cisco-open/go-lanai/pkg/vault"
	"github.com/cisco-open/go-lanai/test"
	"github.com/cisco-open/go-lanai/test/apptest"
	"github.com/cisco-open/go-lanai/test/ittest"
	"github.com/hashicorp/vault/api"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
	"sync"
	"testing"
	"time"
)

/*************************
	Tests
 *************************/

const TestNonExpiringRootToken = `hvs.H8NP7lNhGlg4jX21gRWZvOMn`

var TestRefresherProperties = vault.ConnectionProperties{
	Host:   "127.0.0.1",
	Port:   8200,
	Scheme: "http",
}

type TestRefresherDI struct {
	fx.In
	ittest.RecorderDI
}

func TestRefresherWithK8s(t *testing.T) {
	di := TestRefresherDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		//test.Setup(SetupTestConvertV1HttpRecords("testdata/tokenrefresher/TestRefreshToken.yaml", "testdata/TestRefresherWithK8s.httpvcr.yaml")),
		ittest.WithHttpPlayback(t),
		apptest.WithFxOptions(
			fx.Provide(RecordedVaultProvider()),
		),
		apptest.WithDI(&di),
		test.GomegaSubTest(SubTestRefresherWithK8s(&di), "TestRefresherWithK8s"),
	)
}

func TestRefresherWithToken(t *testing.T) {
	di := TestRefresherDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		ittest.WithHttpPlayback(t),
		//ittest.WithHttpPlayback(t, ittest.HttpRecordingMode()),
		apptest.WithFxOptions(
			fx.Provide(RecordedVaultProvider()),
		),
		apptest.WithDI(&di),
		test.SubTestSetup(SubTestCreateTokens(&di)),
		test.GomegaSubTest(SubTestRefresherWithNonRefreshableToken(&di), "TestRefresherWithNonRefreshableToken"),
		test.GomegaSubTest(SubTestRefresherWithStaticToken(&di), "TestRefresherWithStaticToken"),
	)
}

/*************************
	Sub Tests
 *************************/

func SubTestCreateTokens(di *TestRefresherDI) test.SetupFunc {
	var once sync.Once
	return func(ctx context.Context, t *testing.T) (context.Context, error) {
		var err error
		once.Do(func() {
			g := gomega.NewWithT(t)
			p := TestRefresherProperties
			p.Authentication = vault.Token
			p.Token = TestNonExpiringRootToken
			client := NewTestClient(g, p, di.Recorder)

			req := NewCreateTokenRequest("token_short_ttl", 1 * time.Second, true)
			_, err = client.Logical(ctx).Write("auth/token/create", req)
			if err != nil {
				return
			}

			// note recreating token without ttl will fail, we don't care the result for HTTP replaying purpose
			req = NewCreateTokenRequest("token_no_ttl", 0, false)
			_, _ = client.Logical(ctx).Write("auth/token/create", req)
		})
		return ctx, err
	}
}

func SubTestRefresherWithK8s(di *TestRefresherDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		p := TestRefresherProperties
		p.Authentication = vault.Kubernetes
		p.Kubernetes = vault.KubernetesConfig{
			JWTPath: "testdata/k8s-jwt-refresh.txt",
			Role:    "devweb-app",
		}
		client := NewTestClient(g, p, di.Recorder)
		oldToken := client.Token()
		client.AutoRenewToken(ctx)
		time.Sleep(6 * time.Second)
		newToken := client.Token()
		g.Expect(newToken).NotTo(Equal(oldToken), "Token was not refreshed, before: %v, after: %v", oldToken, newToken)
		//g.Expect(refresher.renewer).NotTo(gomega.BeNil(), "Renewer nilled")
		_ = client.Close()
	}
}

func SubTestRefresherWithNonRefreshableToken(di *TestRefresherDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		p := TestRefresherProperties
		p.Authentication = vault.Token
		p.Token = "token_short_ttl"
		client := NewTestClient(g, p, di.Recorder)
		oldToken := client.Token()
		client.AutoRenewToken(ctx)
		time.Sleep(2500 * time.Millisecond)
		newToken := client.Token()
		g.Expect(newToken).To(Equal(oldToken), "Non-refreshable Token was refreshed, before: %v, after: %v", oldToken, newToken)
		//g.Expect(refresher.renewer).NotTo(gomega.BeNil(), "Renewer nilled")
		_ = client.Close()
	}
}

func SubTestRefresherWithStaticToken(di *TestRefresherDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		p := TestRefresherProperties
		p.Authentication = vault.Token
		p.Token = "token_no_ttl"
		client := NewTestClient(g, p, di.Recorder)

		oldToken := client.Token()
		client.AutoRenewToken(ctx)
		time.Sleep(1500 * time.Millisecond)
		newToken := client.Token()
		g.Expect(newToken).To(Equal(oldToken), "Non-refreshable Token was refreshed, before: %v, after: %v", oldToken, newToken)
		//g.Expect(refresher.renewer).NotTo(gomega.BeNil(), "Renewer nilled")
		_ = client.Close()
	}
}

func SubTestRefresherRestart(di *TestRefresherDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		p := TestRefresherProperties
		p.Authentication = vault.Token
		p.Token = "token_short_ttl"
		client := NewTestClient(g, p, di.Recorder)

		oldToken := client.Token()
		client.AutoRenewToken(ctx)
		time.Sleep(1500 * time.Millisecond)
		newToken := client.Token()
		g.Expect(newToken).To(Equal(oldToken), "Non-refreshable Token was refreshed, before: %v, after: %v", oldToken, newToken)
		//g.Expect(refresher.renewer).NotTo(gomega.BeNil(), "Renewer nilled")
		_ = client.Close()
	}
}

/*************************
	Helpers
 *************************/

func NewCreateTokenRequest(name string, ttl time.Duration, renewable bool) *api.TokenCreateRequest {
	return &api.TokenCreateRequest{
		ID:             name,
		Policies:       []string{"root"},
		TTL:            ttl.String(),
		ExplicitMaxTTL: ttl.String(),
		DisplayName:    name,
		Renewable:      &renewable,
		Type:           "service",
	}
}
