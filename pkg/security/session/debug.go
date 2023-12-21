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

package session

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"net/http"
)

type DebugAuthSuccessHandler struct {}

func (h *DebugAuthSuccessHandler) HandleAuthenticationSuccess(
	_ context.Context, _ *http.Request, _ http.ResponseWriter, from, to security.Authentication) {
	logger.Debugf("session knows auth succeeded: from [%v] to [%v]", from, to)
}

type DebugAuthErrorHandler struct {}

func (h *DebugAuthErrorHandler) HandleAuthenticationError(_ context.Context, _ *http.Request, _ http.ResponseWriter, err error) {
	logger.Debugf("session knows auth failed with %v", err.Error())
}
