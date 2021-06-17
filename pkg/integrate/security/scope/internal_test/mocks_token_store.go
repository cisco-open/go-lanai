package internal_test

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"fmt"
)

type mockedTokenStoreReader struct {
	*mockedBase
}

func (r *mockedTokenStoreReader) ReadAuthentication(_ context.Context, tokenValue string, hint oauth2.TokenHint) (oauth2.Authentication, error) {
	if hint != oauth2.TokenHintAccessToken {
		return nil, fmt.Errorf("[Mocked Error] wrong token hint")
	}
	mt, e := r.parseMockedToken(tokenValue)
	if e != nil {
		return nil, fmt.Errorf("[Mocked Error] invalid access token")
	}

	acct, ok := r.accounts.lookup[mt.UName]
	if !ok {
		return nil, fmt.Errorf("[Mocked Error] invalid access token")
	}

	auth := r.newMockedAuth(mt, acct)
	return auth, nil
}

func (r *mockedTokenStoreReader) ReadAccessToken(_ context.Context, value string) (oauth2.AccessToken, error) {
	mt, e := r.parseMockedToken(value)
	if e != nil {
		return nil, fmt.Errorf("[Mocked Error] invalid access token")
	}

	_, ok := r.accounts.lookup[mt.UName]
	if !ok {
		return nil, fmt.Errorf("[Mocked Error] invalid access token")
	}
	return mt, nil
}

func (r *mockedTokenStoreReader) ReadRefreshToken(_ context.Context, _ string) (oauth2.RefreshToken, error) {
	return nil, fmt.Errorf("ReadRefreshToken is not implemented for mocked token store")
}
