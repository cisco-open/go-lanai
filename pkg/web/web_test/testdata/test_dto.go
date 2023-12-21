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

package testdata

import (
	"net/url"
	"strconv"
)

type JsonRequest struct {
	UriVar     string `uri:"var"`
	QueryVar   string `form:"q"`
	HeaderVar  string `header:"X-VAR"`
	JsonString string `json:"string"`
	JsonInt    int    `json:"int"`
}

type Response struct {
	UriVar     string `json:"uri"`
	QueryVar   string `json:"q"`
	HeaderVar  string `json:"header"`
	JsonString string `json:"string"`
	JsonInt    int    `json:"int"`
}

func newResponse(req *JsonRequest) *Response {
	return &Response{
		UriVar:     req.UriVar,
		QueryVar:   req.QueryVar,
		HeaderVar:  req.HeaderVar,
		JsonString: req.JsonString,
		JsonInt:    req.JsonInt,
	}
}

type JsonResponse Response

func newJsonResponse(req *JsonRequest) *JsonResponse {
	return (*JsonResponse)(newResponse(req))
}

type TextResponse Response

func newTextResponse(req *JsonRequest) *TextResponse {
	return (*TextResponse)(newResponse(req))
}

func (r TextResponse) MarshalText() ([]byte, error) {
	values := url.Values{}
	values.Set("uri", r.UriVar)
	values.Set("q", r.QueryVar)
	values.Set("header", r.HeaderVar)
	values.Set("string", r.JsonString)
	values.Set("int", strconv.Itoa(r.JsonInt))
	return []byte(values.Encode()), nil
}

type BytesResponse Response

func newBytesResponse(req *JsonRequest) *BytesResponse {
	return (*BytesResponse)(newResponse(req))
}

func (r BytesResponse) MarshalBinary() ([]byte, error) {
	return TextResponse(r).MarshalText()
}