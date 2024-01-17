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

package cors

import (
    "github.com/gin-gonic/gin"
    "github.com/rs/cors"
    "net/http"
)

// Options is a configuration container to setup the CORS middleware.
type Options = cors.Options

type corsWrapper struct {
    *cors.Cors
    optionPassthrough bool
}

// build transforms wrapped cors.Cors handler into Gin middleware.
func (c corsWrapper) build() gin.HandlerFunc {
    return func(ctx *gin.Context) {
        c.HandlerFunc(ctx.Writer, ctx.Request)
        if !c.optionPassthrough &&
            ctx.Request.Method == http.MethodOptions &&
            ctx.GetHeader("Access-Control-Request-Method") != "" {
            // Abort processing next Gin middlewares.
            ctx.AbortWithStatus(http.StatusOK)
        }
    }
}

// New creates a new CORS Gin middleware with the provided options.
func New(options Options) gin.HandlerFunc {
    return corsWrapper{cors.New(options), options.OptionsPassthrough}.build()
}

