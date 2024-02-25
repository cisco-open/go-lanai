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

package redismock

import (
    "github.com/cisco-open/go-lanai/test/mocks/internal"
    "github.com/go-redis/redis/v8"
    "github.com/golang/mock/gomock"
    "github.com/onsi/gomega"
    "testing"
)

// TestMockUniversalClient test generated mocks in following aspects:
// - make sure the mock implements the intended interfaces
// - make sure implemented methods doesn't panic
// - make sure implemented methods invocation can be mocked and recorded
func TestMockUniversalClient(t *testing.T) {
    g := gomega.NewWithT(t)
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()
    mock := NewMockUniversalClient(ctrl)

    internal.AssertGoMockGenerated[redis.UniversalClient](g, mock, ctrl)
}
