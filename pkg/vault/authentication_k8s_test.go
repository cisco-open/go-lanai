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
    "github.com/onsi/gomega"
    . "github.com/onsi/gomega"
    "go.uber.org/fx"
    "testing"
)

/*************************
	Test Setup
 *************************/

var TestK8sAuthProperties = vault.ConnectionProperties{
    Host:           "127.0.0.1",
    Port:           8200,
    Scheme:         "http",
    Authentication: "kubernetes",
    Kubernetes:     vault.KubernetesConfig{
        JWTPath: "testdata/k8s-jwt-valid.txt",
        Role:    "devweb-app",
    },
}

/*************************
	Tests
 *************************/

type TestK8sDI struct {
    fx.In
    ittest.RecorderDI
}

func TestAuthenticateWithK8s(t *testing.T) {
    di := TestK8sDI{}
    test.RunTest(context.Background(), t,
        apptest.Bootstrap(),
        test.Setup(SetupTestConvertV1HttpRecords("testdata/authentication_kubernetes/successful_client.yaml", "testdata/TestAuthenticateWithK8s.httpvcr.yaml")),
        ittest.WithHttpPlayback(t),
        apptest.WithFxOptions(
            fx.Provide(RecordedVaultProvider()),
        ),
        apptest.WithDI(&di),
        test.GomegaSubTest(SubTestSuccessfulK8sAuth(&di), "TestSuccessfulK8sAuth"),
    )
}

func TestFailedAuthenticateWithK8s(t *testing.T) {
    di := TestK8sDI{}
    test.RunTest(context.Background(), t,
        apptest.Bootstrap(),
        //test.Setup(SetupTestConvertV1HttpRecords("testdata/authentication_kubernetes/invalid_role.yaml", "testdata/TestFailedAuthenticateWithK8s.httpvcr.yaml")),
        ittest.WithHttpPlayback(t),
        apptest.WithFxOptions(
            fx.Provide(RecordedVaultProvider()),
        ),
        apptest.WithDI(&di),
        test.GomegaSubTest(SubTestFailedK8sAuth(&di), "TestFailedK8sAuth"),
    )
}

/*************************
	Sub-Test Cases
 *************************/

func SubTestSuccessfulK8sAuth(di *TestK8sDI) test.GomegaSubTestFunc {
    return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
        p := TestK8sAuthProperties
        p.Kubernetes.Role = "devweb-app"
        client, e := vault.New(vault.WithProperties(p), VaultWithRecorder(di.Recorder))
        g.Expect(e).To(Succeed(), "client with k8s auth should not fail")
        g.Expect(client).ToNot(BeNil(), "client with k8s auth should not be nil")
        token := client.Token()
        g.Expect(token).ToNot(BeEmpty(), "client's token should not be empty")
    }
}

func SubTestFailedK8sAuth(di *TestK8sDI) test.GomegaSubTestFunc {
    return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
        p := TestK8sAuthProperties
        p.Kubernetes.Role = "invalid-role"
        client, e := vault.New(vault.WithProperties(p), VaultWithRecorder(di.Recorder))
        g.Expect(e).To(Succeed(), "client with k8s auth should not fail")
        g.Expect(client).ToNot(BeNil(), "client with k8s auth should not be nil")
        token := client.Token()
        g.Expect(token).To(BeEmpty(), "client's token should be empty")
    }
}

/*************************
	Helpers
 *************************/
