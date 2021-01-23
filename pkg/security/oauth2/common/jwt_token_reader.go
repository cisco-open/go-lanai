package common

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/jwt"
)

// jwtTokenStoreReader implements TokenStoreReader
type jwtTokenStoreReader struct {
	detailsStore security.ContextDetailsStore
	jwtDecoder jwt.JwtDecoder
}

type JTSROptions func(opt *JTSROption)

type JTSROption struct {
	DetailsStore security.ContextDetailsStore
	Decoder jwt.JwtDecoder
}

func NewJwtTokenStoreReader(opts...JTSROptions) *jwtTokenStoreReader {
	opt := JTSROption{}
	for _, optFunc := range opts {
		optFunc(&opt)
	}

	return &jwtTokenStoreReader{
		detailsStore: opt.DetailsStore,
		jwtDecoder: opt.Decoder,
	}
}

func (r *jwtTokenStoreReader) ReadAuthentication(c context.Context, token oauth2.Token) (oauth2.Authentication, error) {
	panic("implement me")
}

func (r *jwtTokenStoreReader) ReadAccessToken(c context.Context, value string) (oauth2.AccessToken, error) {
	token := oauth2.NewDefaultAccessToken(value)
	// TODO decode JWT
	return token, nil
}

func (r *jwtTokenStoreReader) ReadRefreshToken(c context.Context, value string) (oauth2.RefreshToken, error) {
	token := oauth2.NewDefaultRefreshToken(value)
	// TODO decode JWT
	return token, nil
}
