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

package webtest

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"net/http"
	"net/http/httptest"
)

type TestServerOptions func(conf *TestServerConfig)

type TestServerConfig struct {
	Port           int
	ContextPath    string
	LogLevel       log.LoggingLevel
	RequestOptions []RequestOptions
}

type ExecResult struct {
	Response         *http.Response
	ResponseRecorder *httptest.ResponseRecorder
}
