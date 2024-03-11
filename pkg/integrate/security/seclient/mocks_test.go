package seclient_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"github.com/cisco-open/go-lanai/pkg/security/oauth2"
	"github.com/cisco-open/go-lanai/pkg/security/oauth2/auth"
	"github.com/cisco-open/go-lanai/pkg/web"
	"github.com/cisco-open/go-lanai/pkg/web/rest"
	"net/http"
	"strings"
	"time"
)

const (
	InvalidClientID = "invalid-client"
	ExtClientSecret = "client-secret"
)

type MockedController struct {
	Count int
}

func NewMockedController() *MockedController {
	return &MockedController{}
}

func ProvideWebController(c *MockedController) web.Controller {
	return c
}

func (c *MockedController) Mappings() []web.Mapping {
	return []web.Mapping{
		rest.Post("/v2/token").EndpointFunc(c.Token).Build(),
	}
}

func (c *MockedController) Token(_ context.Context, req *http.Request) (*oauth2.DefaultAccessToken, error) {
	// process request
	tokenReq, e := auth.ParseTokenRequest(req)
	if e != nil {
		return nil, oauth2.NewInvalidTokenRequestError("mocked error")
	}

	if len(tokenReq.ClientId) == 0 {
		tokenReq.ClientId, tokenReq.Extensions[ExtClientSecret] = ExtractBasicAuthInfo(req)
	}

	if len(tokenReq.ClientId) == 0 || tokenReq.ClientId == InvalidClientID {
		return nil, oauth2.NewInvalidClientError("mocked invalid client error")
	}

	// encode request and send it back as access token
	tokenValue := EncodeMockedAccessTokenValue(tokenReq)
	if len(tokenValue) == 0 {
		return nil, oauth2.NewInvalidTokenRequestError("mocked error")
	}
	token := oauth2.NewDefaultAccessToken(tokenValue)
	token.SetExpireTime(time.Now().Add(time.Minute))
	token.SetScopes(tokenReq.Scopes.Copy())
	return token, nil
}

func ExtractBasicAuthInfo(req *http.Request) (clientId, secret string) {
	headerVal := req.Header.Get("Authorization")
	if !strings.HasPrefix(headerVal, "Basic ") {
		return
	}
	b64 := strings.Replace(headerVal, "Basic ", "", 1)
	pair, e := base64.StdEncoding.DecodeString(b64)
	if e != nil {
		return
	}
	split := strings.SplitN(string(pair), ":", 2)
	clientId = split[0]
	if len(split) > 1 {
		secret = split[1]
	}
	return
}

func EncodeMockedAccessTokenValue(tokenReq *auth.TokenRequest) string {
	data, e := json.Marshal(tokenReq)
	if e != nil {
		return ""
	}
	return base64.StdEncoding.EncodeToString(data)
}

func DecodeMockedAccessTokenValue(v string) (*auth.TokenRequest, error) {
	data, e := base64.StdEncoding.DecodeString(v)
	if e != nil {
		return nil, e
	}
	var req auth.TokenRequest
	if e := json.Unmarshal(data, &req); e != nil {
		return nil, e
	}
	return &req, nil
}