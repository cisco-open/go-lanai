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
    "cto-github.cisco.com/NFV-BU/go-lanai/test/ittest"
    "github.com/onsi/gomega"
    . "github.com/onsi/gomega"
    "go.uber.org/fx"
    "gopkg.in/dnaeon/go-vcr.v3/recorder"
    "testing"
)

/*************************
	Tests Setup Helpers
 *************************/

func RecordedVaultProvider() fx.Annotated {
    return fx.Annotated{
        Group: "vault",
        Target: VaultWithRecorder,
    }
}

func VaultWithRecorder(recorder *recorder.Recorder) vault.Options {
    return func(cfg *vault.ClientConfig) error {
        recorder.SetRealTransport(cfg.HttpClient.Transport)
        cfg.HttpClient.Transport = recorder
        return nil
    }
}

func NewTestClient(g *gomega.WithT, props vault.ConnectionProperties, recorder *recorder.Recorder) (*vault.Client,) {
    client, e := vault.New(vault.WithProperties(props), VaultWithRecorder(recorder))
    g.Expect(e).To(Succeed(), "create vault client should not fail")
    return client
}

func SetupTestConvertV1HttpRecords(src, dest string) test.SetupFunc {
    return func(ctx context.Context, t *testing.T) (context.Context, error) {
        e := ittest.ConvertCassetteFileV1toV2(src, dest)
        return ctx, e
    }
}