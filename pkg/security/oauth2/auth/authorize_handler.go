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

package auth

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/session"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/order"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/template"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

const (
	ApprovalModelKeyAuthRequest = "AuthRequest"
	ApprovalModelKeyApprovalUrl = "ApprovalUrl"
)

type ResponseHandlerFunc func(ctx *gin.Context)

type AuthorizeHandler interface {
	// HandleApproved makes various ResponseHandlerFunc of authorization based on
	// 	- response_type
	// 	- scope
	// 	- other parameters
	// if the implementation decide to not to handle the AuthorizeRequest, returns nil, nil.
	// e.g. OIDC impl don't handle non OIDC request and don't handle "code" response type because it's identical from default oauth2 impl
	HandleApproved(ctx context.Context, r *AuthorizeRequest, user security.Authentication) (ResponseHandlerFunc, error)

	// HandleApprovalPage create ResponseHandlerFunc for user approval page
	HandleApprovalPage(ctx context.Context, r *AuthorizeRequest, user security.Authentication) (ResponseHandlerFunc, error)
}

/*************************
	Default
 *************************/

type AuthHandlerOptions func(opt *AuthHandlerOption)

type AuthHandlerOption struct {
	Extensions       []AuthorizeHandler
	ApprovalPageTmpl string
	ApprovalUrl      string
	AuthService      AuthorizationService
	AuthCodeStore    AuthorizationCodeStore
}

// DefaultAuthorizeHandler implements AuthorizeHandler
// it implement standard OAuth2 responses and keep a list of extensions for additional protocols such as OpenID Connect
type DefaultAuthorizeHandler struct {
	extensions       []AuthorizeHandler
	approvalPageTmpl string
	approvalUrl      string
	authService      AuthorizationService
	authCodeStore    AuthorizationCodeStore
}

func NewAuthorizeHandler(opts ...AuthHandlerOptions) *DefaultAuthorizeHandler {
	opt := AuthHandlerOption{
		Extensions:       []AuthorizeHandler{},
		ApprovalPageTmpl: "authorize.tmpl",
	}
	for _, f := range opts {
		f(&opt)
	}

	order.SortStable(opt.Extensions, order.OrderedFirstCompare)
	return &DefaultAuthorizeHandler{
		extensions:       opt.Extensions,
		approvalPageTmpl: opt.ApprovalPageTmpl,
		approvalUrl:      opt.ApprovalUrl,
		authService:      opt.AuthService,
		authCodeStore:    opt.AuthCodeStore,
	}
}

func (h *DefaultAuthorizeHandler) Extend(makers ...AuthorizeHandler) *DefaultAuthorizeHandler {
	h.extensions = append(h.extensions, makers...)
	order.SortStable(h.extensions, order.OrderedFirstCompare)
	return h
}

func (h *DefaultAuthorizeHandler) HandleApproved(ctx context.Context, r *AuthorizeRequest, user security.Authentication) (ResponseHandlerFunc, error) {
	userAuth := ConvertToOAuthUserAuthentication(user)

	// common handling, those common handling could also added as extensions
	h.recordSessionId(ctx, userAuth)

	for _, delegate := range h.extensions {
		if f, e := delegate.HandleApproved(ctx, r, userAuth); f != nil || e != nil {
			return f, e
		}
	}

	switch {
	case r.ResponseTypes.Has("token"):
		return h.MakeImplicitResponse(ctx, r, userAuth)
	case r.ResponseTypes.Has("code"):
		return h.MakeAuthCodeResponse(ctx, r, userAuth)
	default:
		return nil, oauth2.NewInvalidResponseTypeError(fmt.Sprintf("response_type [%v] is not supported", r.ResponseTypes.Values()))
	}
}

func (h *DefaultAuthorizeHandler) HandleApprovalPage(ctx context.Context, r *AuthorizeRequest, user security.Authentication) (ResponseHandlerFunc, error) {
	for _, delegate := range h.extensions {
		if f, e := delegate.HandleApprovalPage(ctx, r, user); f != nil || e != nil {
			return f, e
		}
	}

	return func(ctx *gin.Context) {
		mv := template.ModelView{
			View: h.approvalPageTmpl,
			Model: map[string]interface{}{
				ApprovalModelKeyAuthRequest: r,
				ApprovalModelKeyApprovalUrl: h.approvalUrl,
			},
		}
		_ = template.TemplateEncodeResponseFunc(ctx, ctx.Writer, &mv)
	}, nil
}

func (h *DefaultAuthorizeHandler) MakeAuthCodeResponse(ctx context.Context, r *AuthorizeRequest, user oauth2.UserAuthentication) (ResponseHandlerFunc, error) {
	code, e := h.authCodeStore.GenerateAuthorizationCode(ctx, r, user)
	if e != nil {
		return nil, e
	}

	logger.WithContext(ctx).Debug("authorization_code=" + code)
	values := map[string]string{
		oauth2.ParameterAuthCode: code,
	}

	redirect, e := composeRedirectUrl(ctx, r, values, false)
	if e != nil {
		return nil, e
	}
	return func(c *gin.Context) {
		c.Redirect(http.StatusFound, redirect)
	}, nil
}

func (h *DefaultAuthorizeHandler) MakeImplicitResponse(ctx context.Context, r *AuthorizeRequest, user oauth2.UserAuthentication) (ResponseHandlerFunc, error) {
	//TODO implement Implicit grant
	panic("implicit response is not implemented")
}

/*
************************

		Helpers
	 ************************
*/
func (h *DefaultAuthorizeHandler) recordSessionId(ctx context.Context, user oauth2.UserAuthentication) {
	s := session.Get(ctx)
	if s == nil {
		return
	}
	user.DetailsMap()[security.DetailsKeySessionId] = s.GetID()
}
