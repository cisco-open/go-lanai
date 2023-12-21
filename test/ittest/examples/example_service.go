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

package examples

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/integrate/httpclient"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/integrate/security/scope"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/access"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/rest"
	"net/http"
)

type ExampleRequest struct {
	UseSystemAccount bool   `json:"sysAcct" form:"sys_acct"`
	Username         string `json:"user" form:"user"`
}

type ExampleController struct {
	Service *ExampleService
}

func NewExampleController(svc *ExampleService) web.Controller {
	return &ExampleController{
		Service: svc,
	}
}

func (c *ExampleController) Mappings() []web.Mapping {
	return []web.Mapping{
		rest.Get("/remote").Condition(access.RequirePermissions("DUMMY_PERMISSION")).EndpointFunc(c.Remote).Build(),
		rest.Post("/remote").Condition(access.RequirePermissions("DUMMY_PERMISSION")).EndpointFunc(c.Remote).Build(),
	}
}

func (c *ExampleController) Remote(ctx context.Context, req *ExampleRequest) (interface{}, error) {
	switch {
	case len(req.Username) == 0:
		return c.Service.CallRemoteWithCurrentContext(ctx)
	case req.UseSystemAccount:
		return c.Service.CallRemoteWithSystemAccount(ctx, req.Username)
	default:
		return c.Service.CallRemoteWithoutSystemAccount(ctx, req.Username)
	}
}

type ExampleService struct {
	HttpClient httpclient.Client
}

func NewExampleService(client httpclient.Client) (*ExampleService, error) {
	client, e := client.WithService("usermanagementgoservice")
	if e != nil {
		return nil, e
	}
	return &ExampleService{
		HttpClient: client,
	}, nil
}

// CallRemoteWithSystemAccount switch to given username using system account and make remote HTTP call
func (s *ExampleService) CallRemoteWithSystemAccount(ctx context.Context, username string) (ret interface{}, err error) {
	e := scope.Do(ctx, func(ctx context.Context) {
		ret, err = s.performRemoteHttpCall(ctx)
	}, scope.UseSystemAccount(), scope.WithUsername(username))

	if e != nil {
		return nil, e
	}
	return
}

// CallRemoteWithoutSystemAccount switch to given username directly and make remote HTTP call
func (s *ExampleService) CallRemoteWithoutSystemAccount(ctx context.Context, username string) (ret interface{}, err error) {
	e := scope.Do(ctx, func(ctx context.Context) {
		ret, err = s.performRemoteHttpCall(ctx)
	}, scope.WithUsername(username))

	if e != nil {
		return nil, e
	}
	return
}

// CallRemoteWithCurrentContext make remote HTTP call using current security context
func (s *ExampleService) CallRemoteWithCurrentContext(ctx context.Context) (ret interface{}, err error) {
	return s.performRemoteHttpCall(ctx)
}

func (s *ExampleService) performRemoteHttpCall(ctx context.Context) (ret interface{}, err error) {
	resp, e := s.HttpClient.Execute(ctx, httpclient.NewRequest("/api/v8/users/current", http.MethodGet))
	if e != nil {
		err = e
		return
	}
	ret = resp.Body
	return
}
