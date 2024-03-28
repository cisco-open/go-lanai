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

package web

import (
	"context"
	"encoding"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

/**********************************
	Various Response Encoders
***********************************/

type EncodeOptions func(opt *EncodeOption)

type EncodeOption struct {
	ContentType string
	Writer      http.ResponseWriter
	Response    interface{}
	WriteFunc   func(rw http.ResponseWriter, v interface{}) error
}

func JsonResponseEncoder() EncodeResponseFunc {
	return jsonEncodeResponseFunc
}

func TextResponseEncoder() EncodeResponseFunc {
	return textEncodeResponseFunc
}

func BytesResponseEncoder() EncodeResponseFunc {
	return bytesEncodeResponseFunc
}

func CustomResponseEncoder(opts ...EncodeOptions) EncodeResponseFunc {
	return func(c context.Context, rw http.ResponseWriter, response interface{}) error {
		opts := append([]EncodeOptions{
			func(opt *EncodeOption) {
				opt.Writer = rw
				opt.Response = response
			},
		}, opts...)
		return encodeResponse(c, opts...)
	}
}

/**********************************
	JSON Response Encoder
***********************************/

func JsonWriteFunc(rw http.ResponseWriter, v interface{}) error {
	return json.NewEncoder(rw).Encode(v)
}

func jsonEncodeResponseFunc(c context.Context, rw http.ResponseWriter, response interface{}) error {
	return encodeResponse(c, func(opt *EncodeOption) {
		opt.ContentType = "application/json; charset=utf-8"
		opt.Writer = rw
		opt.Response = response
		opt.WriteFunc = JsonWriteFunc
	})
}

/**********************************
	Text Response Encoder
***********************************/

func TextWriteFunc(rw http.ResponseWriter, v interface{}) error {
	var data []byte
	switch v.(type) {
	case []byte:
		data = v.([]byte)
	case string:
		data = []byte(v.(string))
	case fmt.Stringer:
		data = []byte(v.(fmt.Stringer).String())
	case encoding.TextMarshaler:
		t, e := v.(encoding.TextMarshaler).MarshalText()
		if e != nil {
			return e
		}
		data = t
	default:
		return NewHttpError(http.StatusInternalServerError, errors.New("invalid response type"))
	}
	_, e := rw.Write(data)
	return e
}

func textEncodeResponseFunc(c context.Context, rw http.ResponseWriter, response interface{}) error {
	return encodeResponse(c, func(opt *EncodeOption) {
		opt.ContentType = "text/plain; charset=utf-8"
		opt.Writer = rw
		opt.Response = response
		opt.WriteFunc = TextWriteFunc
	})
}

/**********************************
	Bytes Response Encoder
***********************************/

func BytesWriteFunc(rw http.ResponseWriter, v interface{}) error {
	var data []byte
	switch v.(type) {
	case []byte:
		data = v.([]byte)
	case string:
		data = []byte(v.(string))
	case encoding.BinaryMarshaler:
		t, e := v.(encoding.BinaryMarshaler).MarshalBinary()
		if e != nil {
			return e
		}
		data = t
	default:
		return NewHttpError(http.StatusInternalServerError, errors.New("invalid response type"))
	}
	_, e := rw.Write(data)
	return e
}

func bytesEncodeResponseFunc(c context.Context, rw http.ResponseWriter, response interface{}) error {
	return encodeResponse(c, func(opt *EncodeOption) {
		opt.ContentType = "application/octet-stream"
		opt.Writer = rw
		opt.Response = response
		opt.WriteFunc = BytesWriteFunc
	})
}

/**********************************
	Response Encoding Helpers
***********************************/

// encodeResponse work with endpoint generated with MakeEndpoint
// we could export this function if needed. But for now, it remains hidden
func encodeResponse(_ context.Context, opts ...EncodeOptions) error {
	opt := EncodeOption{}
	for _, f := range opts {
		f(&opt)
	}

	// overwrite headers
	if headerer, ok := opt.Response.(Headerer); ok {
		opt.Writer = NewLazyHeaderWriter(opt.Writer)
		overwriteHeaders(opt.Writer, headerer)
	}

	// additional headers
	opt.Writer.Header().Set("Content-Type", opt.ContentType)

	// write header and status code
	if coder, ok := opt.Response.(StatusCoder); ok {
		opt.Writer.WriteHeader(coder.StatusCode())
	}

	if entity, ok := opt.Response.(BodyContainer); ok {
		opt.Response = entity.Body()
	}
	// we just ignore nil pointer
	switch resp := opt.Response.(type) {
	case nil:
		_, e := opt.Writer.Write([]byte{})
		return e
	default:
		return opt.WriteFunc(opt.Writer, resp)
	}
}

func overwriteHeaders(w http.ResponseWriter, h Headerer) {
	for key, values := range h.Headers() {
		for i, val := range values {
			if i == 0 {
				w.Header().Set(key, val)
			} else {
				w.Header().Add(key, val)
			}
		}
	}
}

