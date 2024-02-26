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

import "github.com/cisco-open/go-lanai/pkg/bootstrap"

type EmptyRequest struct{}

//goland:noinspection GoUnusedConst
const (
	MinWebPrecedence = bootstrap.WebPrecedence
	MaxWebPrecedence = bootstrap.WebPrecedence + bootstrap.FrameworkModulePrecedenceBandwidth

	LowestMiddlewareOrder  = int(^uint(0) >> 1)         // max int
	HighestMiddlewareOrder = -LowestMiddlewareOrder - 1 // min int

	FxGroupControllers     = "controllers"
	FxGroupCustomizers     = "customizers"
	FxGroupErrorTranslator = "error_translators"

	ErrorTemplate = "error.tmpl"

	MethodAny = "ANY"
)

//goland:noinspection GoUnusedConst
const (
	HeaderAuthorization      = "Authorization"
	HeaderOrigin             = "Origin"
	HeaderACAllowCredentials = "Access-Control-Allow-Credentials"
	HeaderACAllowHeaders     = "Access-Control-Allow-Headers"
	HeaderACAllowMethods     = "Access-Control-Allow-AllowedMethodsStr"
	HeaderACAllowOrigin      = "Access-Control-Allow-Origin"
	HeaderACExposeHeaders    = "Access-Control-Expose-Headers"
	HeaderACMaxAge           = "Access-Control-Max-Age"
	HeaderACRequestHeaders   = "Access-Control-Request-Headers"
	HeaderACRequestMethod    = "Access-Control-Request-Method"
	HeaderContentType        = "Content-Type"
	HeaderContentLength      = "Content-Length"
)
