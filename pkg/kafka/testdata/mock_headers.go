package testdata

import (
    "context"
    "cto-github.cisco.com/NFV-BU/go-lanai/pkg/kafka"
    "cto-github.cisco.com/NFV-BU/go-lanai/test"
    "github.com/Shopify/sarama"
    "go.uber.org/fx"
    "sync"
    "testing"
)

type kCtxHeadersMocker struct{}

type MockHeadersOut struct {
    fx.Out
    Concrete *MockedHeadersInterceptor
    Interface kafka.ConsumerDispatchInterceptor `group:"kafka"`
}

func ProvideMockedHeadersInterceptor() MockHeadersOut {
    interceptor := &MockedHeadersInterceptor{
        Headers: make(map[string]map[int32]map[int64]kafka.Headers),
    }
    return MockHeadersOut{
        Concrete:  interceptor,
        Interface: interceptor,
    }
}

type MockHeadersDI struct {
    fx.In
    HeadersMocker *MockedHeadersInterceptor
}

func SubSetupHeadersMocker(di *MockHeadersDI) test.SetupFunc {
    return func(ctx context.Context, t *testing.T) (context.Context, error) {
        if ctx.Value(kCtxHeadersMocker{}) != nil {
            return ctx, nil
        }
        return context.WithValue(ctx, kCtxHeadersMocker{}, di.HeadersMocker), nil
    }
}

func CurrentHeadersMocker(ctx context.Context) *MockedHeadersInterceptor {
    if mocker, ok := ctx.Value(kCtxHeadersMocker{}).(*MockedHeadersInterceptor); ok {
        return mocker
    }
    return &MockedHeadersInterceptor{Headers: make(map[string]map[int32]map[int64]kafka.Headers)}
}

type MockedHeadersInterceptor struct {
    mtx sync.Mutex
    Headers map[string]map[int32]map[int64]kafka.Headers
}

func (i *MockedHeadersInterceptor) Intercept(msgCtx *kafka.MessageContext) (*kafka.MessageContext, error) {
    i.mtx.Lock()
    defer i.mtx.Unlock()
    switch raw := msgCtx.RawMessage.(type) {
    case *sarama.ConsumerMessage:
        partitions, ok := i.Headers[raw.Topic]
        if !ok {
            break
        }
        offsets, ok := partitions[raw.Partition]
        if !ok {
            break
        }
        headers, ok := offsets[raw.Offset]
        if !ok {
            break
        }
        if msgCtx.Message.Headers == nil {
            msgCtx.Message.Headers = make(kafka.Headers)
        }
        for k, v := range headers {
            msgCtx.Message.Headers[k] = v
            raw.Headers = append(raw.Headers, &sarama.RecordHeader{
                Key:   []byte(k),
                Value: []byte(v),
            })
        }
    }
    return msgCtx, nil
}

func (i *MockedHeadersInterceptor) MockHeaders(topic string, partition int32, offset int64, headers kafka.Headers) {
    i.mtx.Lock()
    defer i.mtx.Unlock()
    partitions, ok := i.Headers[topic]
    defer func() {i.Headers[topic] = partitions}()
    if !ok {
        partitions = make(map[int32]map[int64]kafka.Headers)
    }

    offsets, ok := partitions[partition]
    defer func() {partitions[partition] = offsets}()
    if !ok {
        offsets = make(map[int64]kafka.Headers)
    }
    offsets[offset] = headers
}
