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

package logout

import (
    "context"
    "errors"
    "fmt"
    "github.com/cisco-open/go-lanai/pkg/security"
    "github.com/gin-gonic/gin"
)

var ctxKeyWarnings = "logout.Warnings"

func GetWarnings(ctx context.Context) Warnings {
	w, _ := ctx.Value(ctxKeyWarnings).(Warnings)
	return w
}

//goland:noinspection GoNameStartsWithPackageName
type LogoutMiddleware struct {
	successHandler      security.AuthenticationSuccessHandler
	errorHandler        security.AuthenticationErrorHandler
	entryPoint          security.AuthenticationEntryPoint
	logoutHandlers      []LogoutHandler
	conditionalHandlers []ConditionalLogoutHandler
}

func NewLogoutMiddleware(successHandler security.AuthenticationSuccessHandler,
	errorHandler security.AuthenticationErrorHandler,
	entryPoint security.AuthenticationEntryPoint,
	logoutHandlers ...LogoutHandler) *LogoutMiddleware {

	conditionalHandlers := make([]ConditionalLogoutHandler, 0, len(logoutHandlers))
	for _, h := range logoutHandlers {
		if conditional, ok := h.(ConditionalLogoutHandler); ok {
			conditionalHandlers = append(conditionalHandlers, conditional)
		}
	}
	return &LogoutMiddleware{
		successHandler:      successHandler,
		errorHandler:        errorHandler,
		entryPoint:          entryPoint,
		logoutHandlers:      logoutHandlers,
		conditionalHandlers: conditionalHandlers,
	}
}

func (mw *LogoutMiddleware) LogoutHandlerFunc() gin.HandlerFunc {
	return func(gc *gin.Context) {
		before := security.Get(gc)
		// pre-logout check
		for _, h := range mw.conditionalHandlers {
			switch e := h.ShouldLogout(gc, gc.Request, gc.Writer, before); {
			case e != nil && mw.entryPoint != nil:
				mw.handleCancelled(gc, e)
				return
			case e != nil:
				mw.handleError(gc, e)
				return
			}
		}

		// perform logout
		for _, handler := range mw.logoutHandlers {
			switch e := handler.HandleLogout(gc, gc.Request, gc.Writer, before); {
			case errors.Is(e, security.ErrorSubTypeAuthWarning):
				mw.handleWarnings(gc, e)
			case e != nil:
				mw.handleError(gc, e)
				return
			}
		}
		mw.handleSuccess(gc, before)
	}
}

func (mw *LogoutMiddleware) handleSuccess(gc *gin.Context, before security.Authentication) {
	mw.successHandler.HandleAuthenticationSuccess(gc, gc.Request, gc.Writer, before, security.Get(gc))
	if gc.Writer.Written() {
		gc.Abort()
	}
}

func (mw *LogoutMiddleware) handleWarnings(gc *gin.Context, warning error) {
	var warnings Warnings
	existing := gc.Value(ctxKeyWarnings)
	switch v := existing.(type) {
	case Warnings:
		warnings = append(v, warning)
	case []error:
		warnings = append(v, warning)
	case nil:
		warnings = Warnings{warning}
	default:
		warnings = Warnings{fmt.Errorf("%v", existing), warning}
	}
	gc.Set(ctxKeyWarnings, warnings)
}

func (mw *LogoutMiddleware) handleError(gc *gin.Context, err error) {
	if !errors.Is(err, security.ErrorTypeSecurity) {
		err = security.NewInternalAuthenticationError(err.Error(), err)
	}

	mw.errorHandler.HandleAuthenticationError(gc, gc.Request, gc.Writer, err)
	if gc.Writer.Written() {
		gc.Abort()
	}
}

func (mw *LogoutMiddleware) handleCancelled(ctx *gin.Context, err error) {
	mw.entryPoint.Commence(ctx, ctx.Request, ctx.Writer, err)
	if ctx.Writer.Written() {
		ctx.Abort()
	}
}
