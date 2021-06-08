package seclient

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/integrate/httpclient"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"math/rand"
	"net/http"
	"net/url"
	"time"
)

const (
	nonceCharset ="0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

type AuthClientOptions func(opt *AuthClientOption)

type AuthClientOption struct {
	Client            httpclient.Client
	ServiceName       string
	BaseUrl           string
	PwdLoginPath      string
	SwitchContextPath string
	ClientId          string
	ClientSecret      string
}

type remoteAuthClient struct {
	client        httpclient.Client
	clientId      string
	clientSecret  string
	pwdLoginPath  string
	switchCtxPath string
	nonceSeed     *rand.Rand
}

func NewRemoteAuthClient(opts ...AuthClientOptions) *remoteAuthClient {
	opt := AuthClientOption{
		PwdLoginPath: "/v2/token",
		SwitchContextPath: "/v2/token",
	}
	for _, fn := range opts {
		fn(&opt)
	}

	// prepare httpclient
	var client httpclient.Client
	var err error
	if opt.BaseUrl != "" {
		client, err = opt.Client.WithBaseUrl(opt.BaseUrl)
	} else {
		client, err = opt.Client.WithService(opt.ServiceName)
	}
	if err != nil {
		panic(err)
	}

	return &remoteAuthClient{
		client: client.WithConfig(&httpclient.ClientConfig{
			// Note: we don't want access token passthrough
			BeforeHooks: []httpclient.BeforeHook{},
			Logger:     logger,
			MaxRetries: 2,
			Timeout:    30 * time.Second,
			Logging: httpclient.LoggingConfig{
				Level:        log.LevelDebug,
				//DetailsLevel: httpclient.LogDetailsLevelMinimum,
				//SanitizeHeaders: utils.NewStringSet(),
			},
		}),
		clientId:      opt.ClientId,
		clientSecret:  opt.ClientSecret,
		pwdLoginPath:  opt.PwdLoginPath,
		switchCtxPath: opt.SwitchContextPath,
		nonceSeed:     rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (c *remoteAuthClient) PasswordLogin(ctx context.Context, opts ...AuthOptions) (*Result, error) {
	opt := c.option(opts)

	nonce := c.generateNonce(10)
	reqOpts := []httpclient.RequestOptions{
		httpclient.WithParam(oauth2.ParameterGrantType, oauth2.GrantTypePassword),
		httpclient.WithParam(oauth2.ParameterUsername, opt.Username),
		httpclient.WithBasicAuth(c.clientId, c.clientSecret),
		httpclient.WithUrlEncodedBody(url.Values{
			oauth2.ParameterPassword: []string{opt.Password},
			oauth2.ParameterNonce: []string{nonce},
		}),
	}
	reqOpts = append(reqOpts, c.reqOptionsForTenancy(opt)...)

	// prepare request
	req := httpclient.NewRequest(c.pwdLoginPath, http.MethodPost, reqOpts...)
	// send request and parse response
	body := oauth2.NewDefaultAccessToken("")
	resp, e := c.client.Execute(ctx, req, httpclient.JsonBody(body))
	return c.handleResponse(resp, e)
}

func (c *remoteAuthClient) SwitchUser(ctx context.Context, opts ...AuthOptions) (*Result, error) {
	opt := c.option(opts)

	nonce := c.generateNonce(10)
	reqOpts := []httpclient.RequestOptions{
		httpclient.WithParam(oauth2.ParameterGrantType, oauth2.GrantTypeSwitchUser),
		httpclient.WithBasicAuth(c.clientId, c.clientSecret),
		httpclient.WithUrlEncodedBody(url.Values{
			oauth2.ParameterAccessToken: []string{opt.AccessToken},
			oauth2.ParameterNonce: []string{nonce},
		}),
	}
	reqOpts = append(reqOpts, c.reqOptionsForSwitchUser(opt)...)
	reqOpts = append(reqOpts, c.reqOptionsForTenancy(opt)...)

	// prepare request
	req := httpclient.NewRequest(c.switchCtxPath, http.MethodPost, reqOpts...)
	// send request and parse response
	body := oauth2.NewDefaultAccessToken("")
	resp, e := c.client.Execute(ctx, req, httpclient.JsonBody(body))
	return c.handleResponse(resp, e)
}

func (c *remoteAuthClient) SwitchTenant(ctx context.Context, opts ...AuthOptions) (*Result, error) {
	opt := c.option(opts)

	nonce := c.generateNonce(10)
	reqOpts := []httpclient.RequestOptions{
		httpclient.WithParam(oauth2.ParameterGrantType, oauth2.GrantTypeSwitchTenant),
		httpclient.WithBasicAuth(c.clientId, c.clientSecret),
		httpclient.WithUrlEncodedBody(url.Values{
			oauth2.ParameterAccessToken: []string{opt.AccessToken},
			oauth2.ParameterNonce: []string{nonce},
		}),
	}
	reqOpts = append(reqOpts, c.reqOptionsForTenancy(opt)...)

	// prepare request
	req := httpclient.NewRequest(c.switchCtxPath, http.MethodPost, reqOpts...)
	// send request and parse response
	body := oauth2.NewDefaultAccessToken("")
	resp, e := c.client.Execute(ctx, req, httpclient.JsonBody(body))
	return c.handleResponse(resp, e)
}

func (c *remoteAuthClient) option(opts []AuthOptions) *AuthOption {
	opt := AuthOption{}
	for _, fn := range opts {
		fn(&opt)
	}
	return &opt
}

func (c *remoteAuthClient) reqOptionsForTenancy(opt *AuthOption) []httpclient.RequestOptions {
	ret := make([]httpclient.RequestOptions, 0, 2)
	if opt.TenantId != "" {
		ret = append(ret, httpclient.WithParam(oauth2.ParameterTenantId, opt.TenantId))
	}
	if opt.TenantName != "" {
		ret = append(ret, httpclient.WithParam(oauth2.ParameterTenantName, opt.TenantName))
	}
	return ret
}

func (c *remoteAuthClient) reqOptionsForSwitchUser(opt *AuthOption) []httpclient.RequestOptions {
	ret := make([]httpclient.RequestOptions, 0, 2)
	if opt.Username != "" {
		ret = append(ret, httpclient.WithParam(oauth2.ParameterSwitchUsername, opt.Username))
	}
	if opt.UserId != "" {
		ret = append(ret, httpclient.WithParam(oauth2.ParameterSwitchUserId, opt.UserId))
	}
	return ret
}

func (c *remoteAuthClient) handleResponse(resp *httpclient.Response, e error) (*Result, error) {
	if e != nil {
		return nil, e
	}

	token := resp.Body.(oauth2.AccessToken)
	return &Result{
		//Request: nil,
		Token:   token,
	}, nil
}

func (c *remoteAuthClient) generateNonce(length int) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = nonceCharset[c.nonceSeed.Intn(len(nonceCharset))]
	}
	return string(b)
}
