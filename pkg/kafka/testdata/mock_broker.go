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

package testdata

import (
    "context"
    "fmt"
    "github.com/IBM/sarama"
    "github.com/cisco-open/go-lanai/test"
    "github.com/cisco-open/go-lanai/test/apptest"
    "go.uber.org/fx"
    "math/rand"
    "testing"
    "time"
)

type kCtxMockedBroker struct{}

// WithMockedBroker test with mocking provided by sarama.MockBroker.
// The approach we take here is using sarama.MockBroker.SetHandlerByMap. Tester typically mock request/response
// via Mock... functions. e.g. MockExistingTopic
// Issue:   sarama.MockBroker would block when certain request type is not mocked, which would cause cascade failure
//	        of subsequent tests. Typical result of missing request/response mock is test hanging until context expires.
//          This issue would happen when upgrading sarama to newer version.
// Solution: All requests can be seen in sarama.MockBroker.SetHandlerByMap. Find out the missing mocks and investigate
//           the reason. If it's due to added feature or change in sarama package, create new mocks accordingly.
//           To create new mocks, add it into MockBroker.defaults() and modify Mock...() functions based on test cases
func WithMockedBroker() test.Options {
    //nolint:gosec // Not security related
    r := rand.New(rand.NewSource(time.Now().UnixNano()))
    cfg := MockedBrokerConfig{
        Port:   0x7fff + r.Intn(0x7fff) + 1,
        Topics: []string{"test.topic"},
    }
    var mock *MockBroker
    return test.WithOptions(
        test.Setup(func(ctx context.Context, t *testing.T) (context.Context, error) {
            mock = NewMockedBroker(t, &cfg)
            return context.WithValue(ctx, kCtxMockedBroker{}, mock), nil
        }),
        apptest.WithDynamicProperties(map[string]apptest.PropertyValuerFunc{
            "kafka.brokers": func(ctx context.Context) interface{} {
                return fmt.Sprintf("localhost:%d", cfg.Port)
            },
        }),
        apptest.WithFxOptions(fx.Provide(func() *MockBroker {
            return mock
        })),
        test.Teardown(func(ctx context.Context, t *testing.T) error {
            if mock, ok := ctx.Value(kCtxMockedBroker{}).(*MockBroker); ok {
                mock.Close()
            }
            return nil
        }),
    )
}

func CurrentMockedBroker(ctx context.Context) *MockBroker {
    mock, _ := ctx.Value(kCtxMockedBroker{}).(*MockBroker)
    return mock
}

type MockedBrokerConfig struct {
    Port   int
    Topics []string
}

func NewMockedBroker(t *testing.T, cfg *MockedBrokerConfig) *MockBroker {
    mock := sarama.NewMockBrokerAddr(t, 0, fmt.Sprintf(`localhost:%d`, cfg.Port))
    ret := &MockBroker{
        MockBroker: mock,
        t:          t,
        Topics:     make(map[string]map[int32]struct{}),
    }
    ret.Reset()
    return ret
}

type MockResponseUpdateFunc func(mr sarama.MockResponse) sarama.MockResponse

func SetOrAppend(newMR sarama.MockResponse) MockResponseUpdateFunc {
    return func(mr sarama.MockResponse) sarama.MockResponse {
        if mr != nil {
            return sarama.NewMockSequence(mr, newMR)
        }
        return newMR
    }
}

type MockBroker struct {
    *sarama.MockBroker
    t      *testing.T
    Mocks  map[string]sarama.MockResponse
    Topics map[string]map[int32]struct{}
}

func (b *MockBroker) Reset() {
    b.Mocks = b.defaults()
    b.MockBroker.SetHandlerByMap(b.Mocks)
    b.Topics = make(map[string]map[int32]struct{})
}

func (b *MockBroker) UpdateMocks(mocks map[string]MockResponseUpdateFunc) {
    for k := range mocks {
        current, _ := b.Mocks[k]
        b.Mocks[k] = mocks[k](current)
    }
    b.MockBroker.SetHandlerByMap(b.Mocks)
}

func (b *MockBroker) AddTopic(topic string, partition int32, append bool) {
    partitions, ok := b.Topics[topic]
    if !ok {
        partitions = map[int32]struct{}{}
        b.Topics[topic] = partitions
    }
    partitions[partition] = struct{}{}
    if append {
        b.appendMetadataResponse()
    } else {
        b.updateMetadataResponse()
    }
}

func (b *MockBroker) defaults() map[string]sarama.MockResponse {
    return map[string]sarama.MockResponse{
        // General
        "HeartbeatRequest": sarama.NewMockHeartbeatResponse(b.t),
        "MetadataRequest": sarama.NewMockMetadataResponse(b.t).
            SetBroker(b.MockBroker.Addr(), b.MockBroker.BrokerID()).
            SetController(b.MockBroker.BrokerID()),
        "ApiVersionsRequest": sarama.NewMockApiVersionsResponse(b.t),
        // For pubsub
        "OffsetRequest": sarama.NewMockOffsetResponse(b.t),
        "FetchRequest":  sarama.NewMockFetchResponse(b.t, 1),
        // For group
        "OffsetFetchRequest":     sarama.NewMockOffsetFetchResponse(b.t),
        "OffsetCommitRequest":    sarama.NewMockOffsetCommitResponse(b.t),
        "FindCoordinatorRequest": sarama.NewMockFindCoordinatorResponse(b.t),
        "JoinGroupRequest": sarama.NewMockSequence(
            sarama.NewMockJoinGroupResponse(b.t).SetError(sarama.ErrOffsetsLoadInProgress),
            sarama.NewMockJoinGroupResponse(b.t).SetGroupProtocol(sarama.RangeBalanceStrategyName),
        ),
        "SyncGroupRequest": sarama.NewMockSyncGroupResponse(b.t).SetError(sarama.ErrOffsetsLoadInProgress),
    }
}

func (b *MockBroker) updateMetadataResponse() {
    b.UpdateMocks(map[string]MockResponseUpdateFunc{
        "MetadataRequest": func(mr sarama.MockResponse) sarama.MockResponse {
            var resp *sarama.MockMetadataResponse
            switch v := mr.(type) {
            case *sarama.MockMetadataResponse:
                resp = v
            case *sarama.MockSequence:
                resp = sarama.NewMockMetadataResponse(b.t).
                    SetBroker(b.MockBroker.Addr(), b.MockBroker.BrokerID()).
                    SetController(b.MockBroker.BrokerID())
            default:
                return mr
            }
            for topic, partitions := range b.Topics {
                for part := range partitions {
                    resp = resp.SetLeader(topic, part, b.BrokerID())
                }
            }
            return resp
        },
    })
}

func (b *MockBroker) appendMetadataResponse() {
    resp := sarama.NewMockMetadataResponse(b.t).
        SetBroker(b.MockBroker.Addr(), b.MockBroker.BrokerID()).
        SetController(b.MockBroker.BrokerID())
    for topic, partitions := range b.Topics {
        for part := range partitions {
            resp = resp.SetLeader(topic, part, b.BrokerID())
        }
    }
    b.UpdateMocks(map[string]MockResponseUpdateFunc{
        "MetadataRequest": SetOrAppend(resp),
    })
}
