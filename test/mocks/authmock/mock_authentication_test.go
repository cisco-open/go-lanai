package authmock

import (
    "cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
    "cto-github.cisco.com/NFV-BU/go-lanai/test/mocks/internal"
    "github.com/golang/mock/gomock"
    "github.com/onsi/gomega"
    "testing"
)

// TestMockAuthentication test generated mocks in following aspects:
// - make sure the mock implements the intended interfaces
// - make sure implemented methods doesn't panic
// - make sure implemented methods invocation can be mocked and recorded
func TestMockAuthentication(t *testing.T) {
    g := gomega.NewWithT(t)
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()
    mock := NewMockAuthentication(ctrl)

    internal.AssertGoMockGenerated[security.Authentication](g, mock, ctrl)
}
