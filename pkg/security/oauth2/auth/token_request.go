package auth

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"net/http"
)

type TokenRequest struct {
	Parameters map[string]string
	ClientId   string
	Scopes     utils.StringSet
	GrantType  string
	Extensions map[string]interface{}
	context    utils.MutableContext
}

func (r *TokenRequest) Context() utils.MutableContext {
	return r.context
}

func (r *TokenRequest) WithContext(ctx context.Context) *TokenRequest {
	r.context = utils.MakeMutableContext(ctx)
	return r
}

func (r *TokenRequest) OAuth2Request(client oauth2.OAuth2Client) oauth2.OAuth2Request {
	return oauth2.NewOAuth2Request(func(details *oauth2.RequestDetails) {
		details.Parameters = r.Parameters
		details.ClientId = client.ClientId()
		details.Scopes = r.Scopes
		details.Approved = true
		details.GrantType = r.GrantType
		details.Extensions = r.Extensions
	})
}

func NewTokenRequest() *TokenRequest {
	return &TokenRequest{
		Parameters:    map[string]string{},
		Scopes:        utils.NewStringSet(),
		Extensions:    map[string]interface{}{},
		context:       utils.NewMutableContext(),
	}
}

func ParseTokenRequest(req *http.Request) (*TokenRequest, error) {
	if err := req.ParseForm(); err != nil {
		return nil, err
	}

	values := flattenValuesToMap(req.Form);
	return &TokenRequest{
		Parameters:    toStringMap(values),
		ClientId:      extractStringParam(oauth2.ParameterClientId, values),
		Scopes:        extractStringSetParam(oauth2.ParameterScope, " ", values),
		GrantType:     extractStringParam(oauth2.ParameterGrantType, values),
		Extensions:    values,
		context:       utils.MakeMutableContext(req.Context()),
	}, nil
}
