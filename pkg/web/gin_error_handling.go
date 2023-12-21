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
	"github.com/gin-gonic/gin"
	"net/http"
)

/**************************
	Support GenHandling
 **************************/

// GinErrorHandlingCustomizer implements Customizer
type GinErrorHandlingCustomizer struct {

}

func NewGinErrorHandlingCustomizer() *GinErrorHandlingCustomizer {
	return &GinErrorHandlingCustomizer{}
}

func (c GinErrorHandlingCustomizer) Customize(ctx context.Context, r *Registrar) error {
	return r.AddGlobalMiddlewares(DefaultErrorHandling())
}

// DefaultErrorHandling implement error handling logics at last resort, in case errors are not properly handled downstream
func DefaultErrorHandling() gin.HandlerFunc {
	return func(gc *gin.Context) {
		gc.Next()

		if gc.Writer.Written() || len(gc.Errors) == 0 {
			return
		}

		// find first error that implements StatusCoder
		// if not found, use the first one
		err := gc.Errors[0].Err
		for _, e := range gc.Errors {
			//nolint:errorlint
			if _,ok := e.Err.(StatusCoder); !ok {
				err = e.Err
				break
			}
		}
		handleError(gc.Request.Context(), err, gc.Writer)
	}
}

//nolint:errorlint
func handleError(_ context.Context, err error, rw http.ResponseWriter) {
	// body
	contentType, body := "text/plain; charset=utf-8", []byte{}

	// prefer JSON if available
	if marshaler, ok := err.(json.Marshaler); len(body) == 0 && ok {
		if jsonBody, e := marshaler.MarshalJSON(); e == nil {
			contentType, body = "application/json; charset=utf-8", jsonBody
		}
	}
	// then try text
	if marshaler, ok := err.(encoding.TextMarshaler); len(body) == 0 && ok {
		if textBody, e := marshaler.MarshalText(); e == nil {
			body = textBody
		}
	}

	if len(body) == 0 {
		body = []byte(err.Error())
	}


	// header
	rw.Header().Set("Content-Type", contentType)
	if headerer, ok := err.(Headerer); ok {
		for k, values := range headerer.Headers() {
			for _, v := range values {
				rw.Header().Add(k, v)
			}
		}
	}

	// status code
	code := http.StatusInternalServerError
	if sc, ok := err.(StatusCoder); ok {
		code = sc.StatusCode()
	}
	rw.WriteHeader(code)
	_, _ = rw.Write(body)
}
