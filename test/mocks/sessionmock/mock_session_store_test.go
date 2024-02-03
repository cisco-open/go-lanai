package sessionmock

import (
    "cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/session"
    "cto-github.cisco.com/NFV-BU/go-lanai/test/mocks/internal"
    "github.com/golang/mock/gomock"
    "github.com/onsi/gomega"
    "testing"
)

// TestMockStore test generated mocks in following aspects:
// - make sure the mock implements the intended interfaces
// - make sure implemented methods doesn't panic
// - make sure implemented methods invocation can be mocked and recorded
func TestMockStore(t *testing.T) {
	g := gomega.NewWithT(t)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mock := NewMockStore(ctrl)

	internal.AssertGoMockGenerated[session.Store](g, mock, ctrl)
}
