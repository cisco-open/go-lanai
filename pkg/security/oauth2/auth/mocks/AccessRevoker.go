// Code generated by mockery v2.20.0. DO NOT EDIT.

package mocks

import (
	context "context"

	auth "github.com/cisco-open/go-lanai/pkg/security/oauth2/auth"

	mock "github.com/stretchr/testify/mock"
)

// AccessRevoker is an autogenerated mock type for the AccessRevoker type
type AccessRevoker struct {
	mock.Mock
}

// RevokeWithClientId provides a mock function with given fields: ctx, clientId, revokeRefreshToken
func (_m *AccessRevoker) RevokeWithClientId(ctx context.Context, clientId string, revokeRefreshToken bool) error {
	ret := _m.Called(ctx, clientId, revokeRefreshToken)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, bool) error); ok {
		r0 = rf(ctx, clientId, revokeRefreshToken)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// RevokeWithSessionId provides a mock function with given fields: ctx, sessionId, sessionName
func (_m *AccessRevoker) RevokeWithSessionId(ctx context.Context, sessionId string, sessionName string) error {
	ret := _m.Called(ctx, sessionId, sessionName)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) error); ok {
		r0 = rf(ctx, sessionId, sessionName)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// RevokeWithTokenValue provides a mock function with given fields: ctx, tokenValue, hint
func (_m *AccessRevoker) RevokeWithTokenValue(ctx context.Context, tokenValue string, hint auth.RevokerTokenHint) error {
	ret := _m.Called(ctx, tokenValue, hint)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, auth.RevokerTokenHint) error); ok {
		r0 = rf(ctx, tokenValue, hint)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// RevokeWithUsername provides a mock function with given fields: ctx, username, revokeRefreshToken
func (_m *AccessRevoker) RevokeWithUsername(ctx context.Context, username string, revokeRefreshToken bool) error {
	ret := _m.Called(ctx, username, revokeRefreshToken)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, bool) error); ok {
		r0 = rf(ctx, username, revokeRefreshToken)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

type mockConstructorTestingTNewAccessRevoker interface {
	mock.TestingT
	Cleanup(func())
}

// NewAccessRevoker creates a new instance of AccessRevoker. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewAccessRevoker(t mockConstructorTestingTNewAccessRevoker) *AccessRevoker {
	mock := &AccessRevoker{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
