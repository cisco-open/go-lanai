// Copyright 2023 Cisco Systems, Inc. and its affiliates
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package seclient

import (
	"context"
	"github.com/cisco-open/go-lanai/pkg/integrate/httpclient"
	"github.com/cisco-open/go-lanai/pkg/log"
	"github.com/cisco-open/go-lanai/pkg/security/oauth2"
	"github.com/cisco-open/go-lanai/pkg/utils"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	nonceCharset = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

type AuthClientOptions func(opt *AuthClientOption)

type AuthClientOption struct {
	Client            httpclient.Client
	ServiceName       string
	Scheme            string
	ContextPath       string
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
}

func NewRemoteAuthClient(opts ...AuthClientOptions) (AuthenticationClient, error) {
	opt := AuthClientOption{
		PwdLoginPath:      "/v2/token",
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
		client, err = opt.Client.WithService(opt.ServiceName, func(sdOpt *httpclient.SDOption) {
			sdOpt.Scheme = opt.Scheme
			sdOpt.ContextPath = opt.ContextPath
		})
	}
	if err != nil {
		return nil, err
	}

	return &remoteAuthClient{
		client: client.WithConfig(&httpclient.ClientConfig{
			// Note: we don't want access token passthrough
			BeforeHooks: []httpclient.BeforeHook{},
			Logger:      logger,
			MaxRetries:  2,
			Timeout:     30 * time.Second,
			Logging: httpclient.LoggingConfig{
				Level: log.LevelDebug,
				//DetailsLevel: httpclient.LogDetailsLevelMinimum,
				//SanitizeHeaders: utils.NewStringSet(),
			},
		}),
		clientId:      opt.ClientId,
		clientSecret:  opt.ClientSecret,
		pwdLoginPath:  opt.PwdLoginPath,
		switchCtxPath: opt.SwitchContextPath,
	}, nil
}

func (c *remoteAuthClient) ClientCredentials(ctx context.Context, opts ...AuthOptions) (*Result, error) {
	opt := c.option(opts)
	reqOpts := []httpclient.RequestOptions{
		c.withClientAuth(opt),
		httpclient.WithHeader(httpclient.HeaderContentType, httpclient.MediaTypeFormUrlEncoded),
		httpclient.WithUrlEncodedBody(WithNonEmptyURLValues(url.Values{
			oauth2.ParameterGrantType: {oauth2.GrantTypeClientCredentials},
			oauth2.ClaimScope:         {strings.Join(opt.Scopes, " ")},
		})),
	}

	reqOpts = append(reqOpts, c.reqOptionsForTenancy(opt)...)

	// prepare request
	req := httpclient.NewRequest(c.pwdLoginPath, http.MethodPost, reqOpts...)
	// send request and parse response
	body := oauth2.NewDefaultAccessToken("")
	resp, e := c.client.Execute(ctx, req, httpclient.JsonBody(body))
	return c.handleResponse(resp, e)
}

func (c *remoteAuthClient) PasswordLogin(ctx context.Context, opts ...AuthOptions) (*Result, error) {
	opt := c.option(opts)

	nonce := c.generateNonce(10)
	reqOpts := []httpclient.RequestOptions{
		httpclient.WithParam(oauth2.ParameterGrantType, oauth2.GrantTypePassword),
		httpclient.WithParam(oauth2.ParameterUsername, opt.Username),
		c.withClientAuth(opt),
		httpclient.WithUrlEncodedBody(WithNonEmptyURLValues(url.Values{
			oauth2.ParameterPassword: {opt.Password},
			oauth2.ParameterNonce:    {nonce},
			oauth2.ClaimScope:        {strings.Join(opt.Scopes, " ")},
		})),
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
		c.withClientAuth(opt),
		httpclient.WithUrlEncodedBody(WithNonEmptyURLValues(url.Values{
			oauth2.ParameterAccessToken: {opt.AccessToken},
			oauth2.ParameterNonce:       {nonce},
			oauth2.ClaimScope:           {strings.Join(opt.Scopes, " ")},
		})),
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
		c.withClientAuth(opt),
		httpclient.WithUrlEncodedBody(WithNonEmptyURLValues(url.Values{
			oauth2.ParameterAccessToken: {opt.AccessToken},
			oauth2.ParameterNonce:       {nonce},
			oauth2.ClaimScope:           {strings.Join(opt.Scopes, " ")},
		})),
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
	if opt.TenantExternalId != "" {
		ret = append(ret, httpclient.WithParam(oauth2.ParameterTenantExternalId, opt.TenantExternalId))
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
		Token: token,
	}, nil
}

func (c *remoteAuthClient) generateNonce(length int) string {
	return utils.RandomString(length)
}

// withClientAuth will return a requestOption based off of WithBasicAuth, but
// use the clientID from the AuthOptions. If the AuthOption.ClientID is empty, then
// it will return WithBasicAuth using the fallback remoteAuthClient.clientId and secret instead
func (c *remoteAuthClient) withClientAuth(opt *AuthOption) httpclient.RequestOptions {
	clientID := opt.ClientID
	secret := opt.ClientSecret

	if clientID == "" {
		clientID = c.clientId
		secret = c.clientSecret
	}

	return httpclient.WithBasicAuth(clientID, secret)
}

// WithNonEmptyURLValues will accept a map[key][values] and convert it to a url.Values.
// The function will check that the values, typed []string has a length > 0. Otherwise,
// will not insert the key into the url.Values
func WithNonEmptyURLValues(mappedValues map[string][]string) url.Values {
	urlValues := url.Values{}
	for valueKey, values := range mappedValues {
		var nonEmptyValues []string
		for _, value := range values {
			if value != "" {
				nonEmptyValues = append(nonEmptyValues, value)
			}
		}
		if len(nonEmptyValues) > 0 {
			urlValues[valueKey] = nonEmptyValues
		}
	}
	return urlValues
}
