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
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/vault"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/apptest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/ittest"
	"github.com/hashicorp/vault/api"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
	"testing"
	"time"
)

var TestConnProperties = vault.ConnectionProperties{
	Host:           "127.0.0.1",
	Port:           8200,
	Scheme:         "http",
	Authentication: vault.Token,
	Token:          "replace_with_token_value",
}

/*************************
	Tests
 *************************/

//func TestMain(m *testing.M) {
//	suitetest.RunTests(m,
//		ittest.PackageHttpRecordingMode(),
//	)
//}

type TestConnDI struct {
	fx.In
	ittest.RecorderDI
}

func TestConnection(t *testing.T) {
	di := TestConnDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		ittest.WithHttpPlayback(t),
		apptest.WithFxOptions(
			fx.Provide(RecordedVaultProvider()),
		),
		apptest.WithDI(&di),
		test.GomegaSubTest(SubTestNewClient(&di), "TestNewClient"),
		test.GomegaSubTest(SubTestCloneClient(&di), "TestCloneClient"),
	)
}

/*************************
	Sub-Test Cases
 *************************/

func SubTestNewClient(di *TestConnDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		p := TestConnProperties
		hook := &TestHook{}
		client, e := vault.New(vault.WithProperties(p), VaultWithRecorder(di.Recorder))
		g.Expect(e).To(Succeed(), "new client should not fail")
		client.AddHooks(ctx, hook)
		AssertClient(ctx, g, client, hook)
	}
}

func SubTestCloneClient(di *TestConnDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		// initial client
		p := TestConnProperties
		hook := &TestHook{}
		client, e := vault.New(vault.WithProperties(p), VaultWithRecorder(di.Recorder))
		g.Expect(e).To(Succeed(), "new client should not fail")
		client.AddHooks(ctx, hook)

		// create a short lived token
		req := NewCreateTokenRequest("token_short_ttl", 5 * time.Second, false)
		_, e = client.Logical(ctx).Write("auth/token/create", req)
		g.Expect(e).To(Succeed(), "create a temp token should not fail")

		// Clone a client with the new token
		p.Token = "token_short_ttl"
		cloned, e := client.Clone(vault.WithProperties(p))
		g.Expect(e).To(Succeed(), "cloning client should not fail")
		AssertClient(ctx, g, cloned, hook)
	}
}

/*************************
	Helpers
 *************************/

var TestPayload = map[string]interface{}{
	"message": "hello",
	"ttl": "10s",
}

func AssertClient(ctx context.Context, g *gomega.WithT, client *vault.Client, hook *TestHook) {
	g.Expect(client).ToNot(BeNil(), "client should not be nil")
	g.Expect(client.Token()).ToNot(BeEmpty(), "client should have token")
	var sec *api.Secret
	var e error
	// Logical
	sec, e = client.Logical(ctx).Write("/secret/test/foo", TestPayload)
	AssertSecret(g, sec, e, "Write", hook)
	sec, e = client.Logical(ctx).Read("/secret/test/foo")
	AssertSecret(g, sec, e, "Read", hook)
	// Sys
	resp, e := client.Sys(ctx).SealStatus()
	AssertSysResponse(g, resp, e, "Read", hook)
}

func AssertSecret(g *gomega.WithT, sec *api.Secret, err error, op string, hook *TestHook) {
	g.Expect(err).To(Succeed(), "%s should not fail", op)
	if op == "Read" {
		g.Expect(sec).ToNot(BeNil(), "%s should return non-nil result", op)
		g.Expect(sec.Data).ToNot(BeNil(), "%s should return non-nil result", op)
	}
	g.Expect(hook.LastCmd()).To(HavePrefix(op), "hook should registered command [%s ...]", op)
}

func AssertSysResponse[T any](g *gomega.WithT, resp *T, err error, op string, _ *TestHook) {
	g.Expect(err).To(Succeed(), "%s should not fail", op)
	g.Expect(resp).ToNot(BeNil(), "%s should return non-nil result", op)
	// Currently, hook is not instrumenting Sys() ops
}

type TestHook struct {
	Cmds   []string
	Errors []error
}

func (h *TestHook) BeforeOperation(ctx context.Context, cmd string) context.Context {
	h.Cmds = append(h.Cmds, cmd)
	return ctx
}

func (h *TestHook) AfterOperation(_ context.Context, err error) {
	h.Errors = append(h.Errors, err)
}

func (h *TestHook) LastCmd() string {
	if len(h.Cmds) == 0 {
		return ""
	}
	return h.Cmds[len(h.Cmds)-1]
}

