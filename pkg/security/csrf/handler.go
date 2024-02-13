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

package csrf

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"net/http"
)

type ChangeCsrfHandler struct{
	csrfTokenStore TokenStore
}

func (h *ChangeCsrfHandler) HandleAuthenticationSuccess(c context.Context, _ *http.Request, _ http.ResponseWriter, from, to security.Authentication) {
	if !security.IsBeingAuthenticated(from, to) {
		return
	}

	// TODO: review error handling of this block
	t, err := h.csrfTokenStore.LoadToken(c)

	if err != nil {
		panic(security.NewInternalError(err.Error()))
	}

	if t != nil {
		t = h.csrfTokenStore.Generate(c, t.ParameterName, t.HeaderName)
		if e := h.csrfTokenStore.SaveToken(c, t); e != nil {
			panic(security.NewInternalError(err.Error()))
		}
	}

	if mc := utils.FindMutableContext(c); mc != nil {
		mc.Set(web.ContextKeyCsrf, t)
	}

	if gc := web.GinContext(c); gc != nil {
		gc.Set(web.ContextKeyCsrf, t)
	}
}