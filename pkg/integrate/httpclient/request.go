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

package httpclient

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// CreateRequestFunc is a function to create http.Request with given context, method and target URL
type CreateRequestFunc func(ctx context.Context, method string, target *url.URL) (*http.Request, error)

// EncodeRequestFunc is a function to modify http.Request for encoding given value
type EncodeRequestFunc func(ctx context.Context, req *http.Request, val interface{}) error

// RequestOptions used to configure Request in NewRequest
type RequestOptions func(r *Request)

// Request is wraps all information about the request
type Request struct {
	Path           string
	Method         string
	Params         map[string]string
	Headers        http.Header
	Body           interface{}
	BodyEncodeFunc EncodeRequestFunc
	CreateFunc     CreateRequestFunc
}

func NewRequest(path, method string, opts ...RequestOptions) *Request {
	r := Request{
		Path:           path,
		Method:         method,
		Params:         map[string]string{},
		Headers:        http.Header{},
		BodyEncodeFunc: EncodeJSONRequestBody,
		CreateFunc:     defaultRequestCreateFunc,
	}
	for _, f := range opts {
		f(&r)
	}
	return &r
}

func (r Request) encodeHTTPRequest(ctx context.Context, req *http.Request) error {
	// set headers
	for k := range r.Headers {
		req.Header.Set(k, r.Headers.Get(k))
	}

	// set params
	r.applyParams(req)

	return r.BodyEncodeFunc(ctx, req, r.Body)
}

func (r Request) applyParams(req *http.Request) {
	if len(r.Params) == 0 {
		return
	}

	queries := make([]string, len(r.Params))
	i := 0
	for k, v := range r.Params {
		queries[i] = k + "=" + url.QueryEscape(v)
		i++
	}
	req.URL.RawQuery = strings.Join(queries, "&")
}

/**********************
	Defaults
 **********************/

func EncodeJSONRequestBody(_ context.Context, r *http.Request, body interface{}) error {
	if body == nil {
		r.Body = nil
		r.GetBody = nil
		r.ContentLength = 0
		return nil
	}

	if len(r.Header.Values(HeaderContentType)) == 0 {
		r.Header.Set(HeaderContentType, MediaTypeJson)
	}

	var b bytes.Buffer
	r.Body = io.NopCloser(&b)
	err := json.NewEncoder(&b).Encode(body)
	if err != nil {
		return err
	}

	buf := b.Bytes()
	r.GetBody = func() (io.ReadCloser, error) {
		r := bytes.NewReader(buf)
		return io.NopCloser(r), nil
	}
	r.ContentLength = int64(b.Len())
	return nil
}

func EncodeURLEncodedRequestBody(_ context.Context, r *http.Request, body interface{}) error {
	values, ok := body.(url.Values)
	if !ok {
		return NewRequestSerializationError(fmt.Errorf("www-form-urlencoded body expects url.Values but got %T", body))
	}

	if len(r.Header.Values(HeaderContentType)) == 0 {
		r.Header.Set(HeaderContentType, MediaTypeFormUrlEncoded)
	}

	encoded := values.Encode()
	r.GetBody = func() (io.ReadCloser, error) {
		r := strings.NewReader(encoded)
		return io.NopCloser(r), nil
	}
	r.Body, _ = r.GetBody()
	r.ContentLength = int64(len(encoded))
	return nil
}

func defaultRequestCreateFunc(ctx context.Context, method string, target *url.URL) (*http.Request, error) {
	return http.NewRequestWithContext(ctx, method, target.String(), nil)
}

/**********************
	Request Options
 **********************/

func WithoutHeader(key string) RequestOptions {
	switch {
	case key == "":
		return noop()
	default:
		return func(r *Request) {
			r.Headers.Del(key)
		}
	}
}

func WithHeader(key, value string) RequestOptions {
	switch {
	case key == "" || value == "":
		return noop()
	default:
		return func(r *Request) {
			r.Headers.Add(key, value)
		}
	}
}

func WithParam(key, value string) RequestOptions {
	switch {
	case key == "":
		return noop()
	case value == "":
		return func(r *Request) {
			delete(r.Params, key)
		}
	default:
		return func(r *Request) {
			r.Params[key] = value
		}
	}
}

func WithBody(body interface{}) RequestOptions {
	return func(r *Request) {
		r.Body = body
	}
}

func WithRequestBodyEncoder(enc EncodeRequestFunc) RequestOptions {
	return func(r *Request) {
		r.BodyEncodeFunc = enc
		if r.BodyEncodeFunc == nil {
			r.BodyEncodeFunc = EncodeJSONRequestBody
		}
	}
}

func WithRequestCreator(enc CreateRequestFunc) RequestOptions {
	return func(r *Request) {
		r.CreateFunc = enc
		if r.CreateFunc == nil {
			r.CreateFunc = defaultRequestCreateFunc
		}
	}
}

func WithBasicAuth(username, password string) RequestOptions {
	raw := username + ":" + password
	b64 := base64.StdEncoding.EncodeToString([]byte(raw))
	auth := "Basic " + b64
	return WithHeader(HeaderAuthorization, auth)
}

func WithUrlEncodedBody(body url.Values) RequestOptions {
	return mergeRequestOptions(WithBody(body), WithRequestBodyEncoder(EncodeURLEncodedRequestBody))
}

func mergeRequestOptions(opts...RequestOptions) RequestOptions {
	return func(r *Request) {
		for _, fn := range opts {
			fn(r)
		}
	}
}

func noop() func(r *Request) {
	return func(_ *Request) {
		// noop
	}
}
