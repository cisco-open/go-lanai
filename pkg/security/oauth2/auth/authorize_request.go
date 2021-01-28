package auth

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type AuthorizeRequest struct {
	Parameters    map[string]string
	ClientId      string
	ResponseTypes utils.StringSet
	Scopes        utils.StringSet
	RedirectUri   string
	State         string
	Extensions    map[string]interface{}
	Approved      bool
	context		  utils.MutableContext

	// TODO should we still support resource IDs ?
}

func (r *AuthorizeRequest) Context() utils.MutableContext {
	return r.context
}

func (r *AuthorizeRequest) OAuth2Request() oauth2.OAuth2Request {
	return oauth2.NewOAuth2Request(func(details *oauth2.RequestDetails) {
		if grantType, ok := r.Parameters[oauth2.ParameterGrantType]; ok {
			details.GrantType = grantType
		}

		details.Parameters = r.Parameters
		details.ClientId = r.ClientId
		details.Scopes = r.Scopes
		details.Approved = true
		details.RedirectUri = r.RedirectUri
		details.ResponseTypes = r.ResponseTypes
		details.Extensions = r.Extensions
	})
}

func NewAuthorizeRequest(req *http.Request) *AuthorizeRequest {
	return &AuthorizeRequest{
		Parameters:    map[string]string{},
		ResponseTypes: utils.NewStringSet(),
		Scopes:        utils.NewStringSet(),
		Extensions:    map[string]interface{}{},
		context:       utils.MakeMutableContext(req.Context()),
	}
}

func ParseAuthorizeRequest(req *http.Request) (*AuthorizeRequest, error) {
	if err := req.ParseForm(); err != nil {
		return nil, err
	}

	values := flattenValuesToMap(req.Form);
	return &AuthorizeRequest{
		Parameters:    toStringMap(values),
		ClientId:      extractStringParam(oauth2.ParameterClientId, values),
		ResponseTypes: extractStringSetParam(oauth2.ParameterResponseType, " ", values),
		Scopes:        extractStringSetParam(oauth2.ParameterScope, " ", values),
		RedirectUri:   extractStringParam(oauth2.ParameterResponseType, values),
		State:         extractStringParam(oauth2.ParameterState, values),
		Extensions:    values,
		context:       utils.MakeMutableContext(req.Context()),
	}, nil
}


/************************
	Helpers
 ************************/
func flattenValuesToMap(src url.Values) (dest map[string]interface{}) {
	dest = map[string]interface{}{}
	for k, v := range src {
		if len(v) == 0 {
			continue
		}
		dest[k] = strings.Join(v, " ")
	}
	return
}

func toStringMap(src map[string]interface{}) (dest map[string]string) {
	dest = map[string]string{}
	for k, v := range src {
		switch v.(type) {
		case string:
			dest[k] = v.(string)
		case fmt.Stringer:
			dest[k] = v.(fmt.Stringer).String()
		}
	}
	return
}

func extractStringParam(key string, params map[string]interface{}) string {
	if v, ok := params[key]; ok {
		delete(params, key)
		return v.(string)
	}
	return ""
}

func extractStringSetParam(key, sep string, params map[string]interface{}) utils.StringSet {
	if v, ok := params[key]; ok {
		delete(params, key)
		return utils.NewStringSet(strings.Split(v.(string), sep)...)
	}
	return utils.NewStringSet()
}