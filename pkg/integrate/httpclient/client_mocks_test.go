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

package httpclient_test

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/cisco-open/go-lanai/pkg/web"
	"github.com/cisco-open/go-lanai/pkg/web/rest"
	"io"
	"net/http"
	"strconv"
)

const (
	TestErrorMessageSuffix = " - so what?"
	TestErrorDescription   = "This is a generated error"
)

type ServerEchoResponse struct {
	Headers map[string]string `json:"headers"`
	Form    map[string]string `json:"form"`
	ReqBody json.RawMessage   `json:"body"`
}

type ServerErrorResponse struct {
	ServerEchoResponse
	SC      int `json:"-"`
	Message string
}

func (r ServerErrorResponse) Error() string {
	return r.Message
}

func (r ServerErrorResponse) StatusCode() int {
	return r.SC
}

func (r ServerErrorResponse) MarshalJSON() ([]byte, error) {
	v := struct {
		ServerEchoResponse
		Error   string             `json:"error"`
		Message string             `json:"message"`
		Desc    string             `json:"error_description"`
		Details ServerEchoResponse `json:"details"`
	}{
		ServerEchoResponse: r.ServerEchoResponse,
		Error:              TestErrorMessageSuffix,
		Message:            r.Message,
		Desc:               TestErrorDescription,
		Details:            ServerEchoResponse{},
	}
	return json.Marshal(v)
}

type ServerNoContentResponse struct {
	Headers map[string]string `json:"headers"`
	Form    map[string]string `json:"form"`
}

type NoContentErrorResponse struct {
	SC int `json:"-"`
}

func (r NoContentErrorResponse) Error() string {
	return "error"
}

func (r NoContentErrorResponse) StatusCode() int {
	return r.SC
}

func (r NoContentErrorResponse) MarshalJSON() ([]byte, error) {
	return nil, nil
}

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
		rest.Post("/timeout").EndpointFunc(c.Timeout).Build(),
		rest.Post("/echo").EndpointFunc(c.Echo).Build(),
		rest.Put("/fail").EndpointFunc(c.Fail).Build(),
		rest.Put("/maybe").EndpointFunc(c.Maybe).Build(),
		rest.Post("/nocontent").EndpointFunc(c.NoContent).Build(),
		rest.Put("/nocontentfail").EndpointFunc(c.NoContentFail).Build(),
	}
}

func (c *MockedController) Timeout(ctx context.Context, req *http.Request) (interface{}, error) {
	select {
	case <-ctx.Done():
	}
	return c.echoResponse(req)
}

func (c *MockedController) Echo(_ context.Context, req *http.Request) (interface{}, error) {
	return c.echoResponse(req)
}

func (c *MockedController) Fail(_ context.Context, req *http.Request) (*ServerEchoResponse, error) {
	echo, e := c.echoResponse(req)
	if e != nil {
		return nil, e
	}
	sc, e := strconv.Atoi(req.Form.Get("sc"))
	if e != nil {
		sc = http.StatusInternalServerError
	}
	return nil, &ServerErrorResponse{
		ServerEchoResponse: *echo,
		SC:                 sc,
		Message:            "failed" + TestErrorMessageSuffix,
	}
}

func (c *MockedController) Maybe(_ context.Context, req *http.Request) (*ServerEchoResponse, error) {
	c.Count++
	echo, e := c.echoResponse(req)
	if e != nil {
		return nil, e
	}
	rate, e := strconv.Atoi(req.Form.Get("rate"))
	if e == nil && c.Count%rate == 0 {
		return echo, nil
	}

	sc, e := strconv.Atoi(req.Form.Get("sc"))
	if e != nil {
		sc = http.StatusInternalServerError
	}
	return nil, &ServerErrorResponse{
		ServerEchoResponse: *echo,
		SC:                 sc,
		Message:            fmt.Sprintf("failed because: [count=%d, rate=%d]%s", c.Count, rate, TestErrorMessageSuffix),
	}
}

func (c *MockedController) echoResponse(req *http.Request) (*ServerEchoResponse, error) {
	ret := ServerEchoResponse{
		Headers: map[string]string{},
		Form:    map[string]string{},
	}

	for k := range req.Header {
		ret.Headers[k] = req.Header.Get(k)
	}

	if e := req.ParseForm(); e != nil {
		return nil, e
	}

	for k := range req.Form {
		ret.Form[k] = req.Form.Get(k)
	}

	data, e := io.ReadAll(req.Body)
	if e != nil {
		return nil, e
	}
	ret.ReqBody = data
	if len(ret.ReqBody) == 0 {
		ret.ReqBody = nil
	}

	return &ret, nil
}

func (c *MockedController) NoContent(_ context.Context, req *http.Request) (int, interface{}, error) {
	ret, err := c.noContentResponse(req)
	if err != nil {
		return 500, nil, err
	}
	return http.StatusNoContent, &ret, nil
}

func (c *MockedController) NoContentFail(_ context.Context, req *http.Request) (*ServerNoContentResponse, error) {
	_, e := c.noContentResponse(req)
	if e != nil {
		return nil, e
	}
	sc, e := strconv.Atoi(req.Form.Get("sc"))
	if e != nil {
		sc = http.StatusInternalServerError
	}
	return nil, &NoContentErrorResponse{
		SC: sc,
	}
}

func (c *MockedController) noContentResponse(req *http.Request) (*ServerNoContentResponse, error) {
	ret := ServerNoContentResponse{
		Headers: map[string]string{},
		Form:    map[string]string{},
	}

	for k := range req.Header {
		ret.Headers[k] = req.Header.Get(k)
	}

	if e := req.ParseForm(); e != nil {
		return nil, e
	}

	for k := range req.Form {
		ret.Form[k] = req.Form.Get(k)
	}
	return &ret, nil
}
