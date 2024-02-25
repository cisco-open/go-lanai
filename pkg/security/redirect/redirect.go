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

package redirect

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/session"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"github.com/gin-gonic/gin"
	"net/http"
	urlutils "net/url"
	"path"
)

const (
	FlashKeyPreviousError      = "error"
	FlashKeyPreviousStatusCode = "status"
)

// RedirectHandler implements multiple interface for authentication and error handling strategies
//goland:noinspection GoNameStartsWithPackageName
type RedirectHandler struct {
	sc            int
	location      string
	ignoreCtxPath bool
}

func NewRedirectWithRelativePath(path string, ignoreCtxPath bool) *RedirectHandler {
	url, err := urlutils.Parse(path)
	if err != nil {
		panic(err)
	}

	return &RedirectHandler{
		sc:       302,
		location: url.String(),
		ignoreCtxPath: ignoreCtxPath,
	}
}

func NewRedirectWithURL(urlStr string) *RedirectHandler {
	url, err := urlutils.Parse(urlStr)
	if err != nil {
		panic(err)
	}

	return &RedirectHandler{
		sc:       302,
		location: url.String(),
	}
}

// Commence implements security.AuthenticationEntryPoint
func (ep *RedirectHandler) Commence(c context.Context, r *http.Request, rw http.ResponseWriter, err error) {
	ep.doRedirect(c, r, rw, map[string]interface{}{
		FlashKeyPreviousError:      err,
		FlashKeyPreviousStatusCode: http.StatusUnauthorized,
	})
}

// HandleAccessDenied implements security.AccessDeniedHandler
func (ep *RedirectHandler) HandleAccessDenied(c context.Context, r *http.Request, rw http.ResponseWriter, err error) {
	ep.doRedirect(c, r, rw, map[string]interface{}{
		FlashKeyPreviousError:      err,
		FlashKeyPreviousStatusCode: http.StatusForbidden,
	})
}

// HandleAuthenticationSuccess implements security.AuthenticationSuccessHandler
func (ep *RedirectHandler) HandleAuthenticationSuccess(c context.Context, r *http.Request, rw http.ResponseWriter, from, to security.Authentication) {
	ep.doRedirect(c, r, rw, nil)
}

// HandleAuthenticationError implements security.AuthenticationErrorHandler
func (ep *RedirectHandler) HandleAuthenticationError(c context.Context, r *http.Request, rw http.ResponseWriter, err error) {
	ep.doRedirect(c, r, rw, map[string]interface{}{
		FlashKeyPreviousError:      err,
		FlashKeyPreviousStatusCode: http.StatusUnauthorized,
	})
}

func (ep *RedirectHandler) doRedirect(c context.Context, r *http.Request, rw http.ResponseWriter, flashes map[string]interface{}) {
	if grw, ok := rw.(gin.ResponseWriter); ok && grw.Written() {
		return
	}

	// save flashes
	if flashes != nil && len(flashes) != 0 {
		s := session.Get(c)
		if s != nil {
			for k, v := range flashes {
				s.AddFlash(v, k)
			}
		}
	}

	location, _ := urlutils.Parse(ep.location)
	if !location.IsAbs() {
		// relative path was used, try to add context path
		contextPath := web.ContextPath(c)
		if !ep.ignoreCtxPath {
			location.Path = path.Join(contextPath, location.Path)
		}
	}

	// redirect
	http.Redirect(rw, r, location.String(), ep.sc)
	_, _ = rw.Write([]byte{})
}
