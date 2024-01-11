package testdata

import (
	"context"
	"github.com/Shopify/sarama"
)

func MockCreateTopic(ctx context.Context, topic string) {
	mock := CurrentMockedBroker(ctx)
	t := mock.t
	updaters := map[string]MockResponseUpdateFunc{
		// TODO how to mock error response?
		"DescribeAclsRequest": SetOrAppend(sarama.NewMockListAclsResponse(t)),
		"CreateTopicsRequest": SetOrAppend(sarama.NewMockCreateTopicsResponse(t)),
	}
	mock.UpdateMocks(updaters)
	mock.AddTopic(topic, 0, true)
}

func MockExistingTopic(ctx context.Context, topic string, partition int32) {
	mock := CurrentMockedBroker(ctx)
	updaters := map[string]MockResponseUpdateFunc{
		"OffsetRequest": func(mr sarama.MockResponse) sarama.MockResponse {
			return mr.(*sarama.MockOffsetResponse).
				SetOffset(topic, partition, sarama.OffsetOldest, 0).
				SetOffset(topic, partition, sarama.OffsetNewest, 0)
		},
	}
	mock.UpdateMocks(updaters)
	mock.AddTopic(topic, partition, false)
}

func MockCreatePartition(ctx context.Context, topic string, partition int32) {
	mock := CurrentMockedBroker(ctx)
	t := mock.t
	updaters := map[string]MockResponseUpdateFunc{
		// TODO how to mock error response?
		"DescribeAclsRequest":     SetOrAppend(sarama.NewMockListAclsResponse(t)),
		"CreatePartitionsRequest": SetOrAppend(sarama.NewMockCreatePartitionsResponse(t)),
	}
	mock.UpdateMocks(updaters)
	mock.AddTopic(topic, partition, true)
}

func MockProduce(ctx context.Context, topic string, fail bool) {
	mock := CurrentMockedBroker(ctx)
	t := mock.t
	resp := sarama.NewMockProduceResponse(t).SetVersion(3)
	if fail {
		resp = resp.SetError(topic, 0, sarama.ErrUnknown)
	} else {
		resp = resp.SetError(topic, 0, sarama.ErrNoError)
	}
	updaters := map[string]MockResponseUpdateFunc{
		"ProduceRequest": SetOrAppend(resp),
	}
	mock.UpdateMocks(updaters)
}

func MockGroup(ctx context.Context, topic string, group string, partition int32) {
	mock := CurrentMockedBroker(ctx)
	updaters := map[string]MockResponseUpdateFunc{
		"FindCoordinatorRequest": func(mr sarama.MockResponse) sarama.MockResponse {
			return mr.(*sarama.MockFindCoordinatorResponse).
				SetCoordinator(sarama.CoordinatorGroup, group, mock.MockBroker)
		},
		"SyncGroupRequest": SetOrAppend(sarama.NewMockSyncGroupResponse(mock.t).SetMemberAssignment(
			&sarama.ConsumerGroupMemberAssignment{
				Version: 3,
				Topics: map[string][]int32{
					topic: {partition},
				},
			}),
		),
		"OffsetCommitRequest": func(mr sarama.MockResponse) sarama.MockResponse {
			return mr.(*sarama.MockOffsetCommitResponse).SetError(group, topic, partition, sarama.ErrNoError)
		},
	}
	mock.UpdateMocks(updaters)
}

func MockSubscribedMessage(ctx context.Context, topic string, partition int32, offset int64, msg MockedMessage) {
	mock := CurrentMockedBroker(ctx)
	updaters := map[string]MockResponseUpdateFunc{
		"OffsetRequest": func(mr sarama.MockResponse) sarama.MockResponse {
			return mr.(*sarama.MockOffsetResponse).
				SetOffset(topic, partition, sarama.OffsetOldest, 0).
				SetOffset(topic, partition, sarama.OffsetNewest, offset)
		},
		"FetchRequest": func(mr sarama.MockResponse) sarama.MockResponse {
			if len(msg.Key) == 0 {
				return mr.(*sarama.MockFetchResponse).SetMessage(topic, partition, offset, sarama.ByteEncoder(msg.Value))
			}
			return mr.(*sarama.MockFetchResponse).SetMessageWithKey(topic, partition, offset, sarama.ByteEncoder(msg.Key), sarama.ByteEncoder(msg.Value))
		},
	}
	mock.UpdateMocks(updaters)

	if len(msg.Headers) != 0 {
		// Mock headers separately
		// Note, as of sarama 1.38.x, the sarama.MockFetchResponse can only add legacy messages (no header supported).
		// See sarama.MockFetchResponse.For(), it uses sarama.FetchResponse.AddMessage instead of sarama.FetchResponse.AddRecord
		CurrentHeadersMocker(ctx).MockHeaders(topic, partition, offset, msg.Headers)
	}
}

func MockGroupMessage(ctx context.Context, topic, group string, partition int32, offset int64, msg MockedMessage) {
	mock := CurrentMockedBroker(ctx)
	updaters := map[string]MockResponseUpdateFunc{
		"OffsetRequest": func(mr sarama.MockResponse) sarama.MockResponse {
			return mr.(*sarama.MockOffsetResponse).
				SetOffset(topic, partition, sarama.OffsetOldest, 0).
				SetOffset(topic, partition, sarama.OffsetNewest, offset)
		},
		"FetchRequest": func(mr sarama.MockResponse) sarama.MockResponse {
			if len(msg.Key) == 0 {
				return mr.(*sarama.MockFetchResponse).SetMessage(topic, partition, offset, sarama.ByteEncoder(msg.Value))
			}
			return mr.(*sarama.MockFetchResponse).SetMessageWithKey(topic, partition, offset, sarama.ByteEncoder(msg.Key), sarama.ByteEncoder(msg.Value))
		},
		"OffsetFetchRequest": func(mr sarama.MockResponse) sarama.MockResponse {
			return mr.(*sarama.MockOffsetFetchResponse).SetOffset(group, topic, partition, offset, "", sarama.ErrNoError)
		},
	}
	mock.UpdateMocks(updaters)

	if len(msg.Headers) != 0 {
		// Mock headers separately
		// Note, as of sarama 1.38.x, the sarama.MockFetchResponse can only add legacy messages (no header supported).
		// See sarama.MockFetchResponse.For(), it uses sarama.FetchResponse.AddMessage instead of sarama.FetchResponse.AddRecord
		CurrentHeadersMocker(ctx).MockHeaders(topic, partition, offset, msg.Headers)
	}
}

type MockedMessage struct {
	Key     []byte
	Value   []byte
	Headers map[string]string
}
