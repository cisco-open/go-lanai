package auth

import (
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

func NewTokenRequest(req *http.Request) *TokenRequest {
	return &TokenRequest{
		Parameters:    map[string]string{},
		Scopes:        utils.NewStringSet(),
		Extensions:    map[string]interface{}{},
		context:       utils.MakeMutableContext(req.Context()),
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
