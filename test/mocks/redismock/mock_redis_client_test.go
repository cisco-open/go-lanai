package redismock

import (
    "cto-github.cisco.com/NFV-BU/go-lanai/test/mocks/internal"
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
