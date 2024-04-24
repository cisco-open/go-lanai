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

package authorize

import (
	"context"
	"errors"
	"fmt"
	"github.com/cisco-open/go-lanai/pkg/security"
	"github.com/cisco-open/go-lanai/pkg/security/oauth2"
	"github.com/cisco-open/go-lanai/pkg/security/oauth2/auth"
	"github.com/cisco-open/go-lanai/pkg/security/session"
	"github.com/cisco-open/go-lanai/pkg/utils"
	"github.com/cisco-open/go-lanai/pkg/web"
	"github.com/gin-gonic/gin"
	"strconv"
	"strings"
)

const (
	sessionKeyAuthorizeRequest = "kAuthorizeRequest"
	sessionKeyApprovedRequests = "kApprovedRequests"
	scopeParamPrefix           = "scope."
)

/***********************
	Authorize Endpoint
 ***********************/

//goland:noinspection GoNameStartsWithPackageName
type AuthorizeEndpointMiddleware struct {
	requestProcessor auth.AuthorizeRequestProcessor
	authorizeHandler auth.AuthorizeHandler
	approveMatcher   web.RequestMatcher
}

//goland:noinspection GoNameStartsWithPackageName
type AuthorizeMWOptions func(*AuthorizeMWOption)

//goland:noinspection GoNameStartsWithPackageName
type AuthorizeMWOption struct {
	RequestProcessor auth.AuthorizeRequestProcessor
	AuthorizeHandler auth.AuthorizeHandler
	ApprovalMatcher  web.RequestMatcher
}

func NewAuthorizeEndpointMiddleware(opts ...AuthorizeMWOptions) *AuthorizeEndpointMiddleware {
	opt := AuthorizeMWOption{
		RequestProcessor: auth.NewAuthorizeRequestProcessor(),
	}
	for _, optFunc := range opts {
		if optFunc != nil {
			optFunc(&opt)
		}
	}
	return &AuthorizeEndpointMiddleware{
		requestProcessor: opt.RequestProcessor,
		authorizeHandler: opt.AuthorizeHandler,
		approveMatcher:   opt.ApprovalMatcher,
	}
}

func (mw *AuthorizeEndpointMiddleware) PreAuthenticateHandlerFunc(condition web.RequestMatcher) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if matches, err := condition.MatchesWithContext(ctx, ctx.Request); !matches || err != nil {
			return
		}

		// parse or load request
		var request *auth.AuthorizeRequest
		var err error
		switch approve, e := mw.approveMatcher.MatchesWithContext(ctx, ctx.Request); {
		case e == nil && approve:
			// approve or deny request
			if request, err = mw.loadAuthorizeRequest(ctx); err != nil {
				err = oauth2.NewInvalidAuthorizeRequestError("error loading authorize request for approval", e)
			}
		default:
			if request, err = auth.ParseAuthorizeRequest(ctx.Request); err != nil {
				err = oauth2.NewInvalidAuthorizeRequestError("invalid authorize request", e)
			}
		}
		if err != nil {
			mw.handleError(ctx, err)
			return
		}

		ctx.Set(oauth2.CtxKeyReceivedAuthorizeRequest, request)

		// validate and process, regardless the result, we might want to transfer some context from request to current context
		processed, e := mw.requestProcessor.Process(ctx, request)
		if e != nil {
			mw.transferContextValues(request.Context(), ctx)
			mw.handleError(ctx, e)
			return
		}

		// everything is ok, set it to context for later usage
		mw.transferContextValues(processed.Context(), ctx)
		ctx.Set(oauth2.CtxKeyValidatedAuthorizeRequest, processed)
	}
}

func (mw *AuthorizeEndpointMiddleware) AuthorizeHandlerFunc(condition web.RequestMatcher) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if matches, err := condition.MatchesWithContext(ctx, ctx.Request); !matches || err != nil {
			return
		}

		// sanity checks
		request, client, user, e := mw.endpointSanityCheck(ctx)
		if e != nil {
			mw.handleError(ctx, e)
			return
		}
		logger.WithContext(ctx).Debug(fmt.Sprintf("AuthorizeRequest: %s", request))

		// check auto-approval and create response
		var respFunc auth.ResponseHandlerFunc
		e = auth.ValidateAllAutoApprovalScopes(ctx, client, request.Scopes)
		needsApproval := false
		if e != nil {
			needsApproval = !mw.authRequestHasSavedApproval(ctx, request)
		}

		if needsApproval {
			// save request
			if e := mw.saveAuthorizeRequest(ctx, request); e != nil {
				mw.handleError(ctx, e)
				return
			}
			respFunc, e = mw.authorizeHandler.HandleApprovalPage(ctx, request, user)
		} else {
			respFunc, e = mw.authorizeHandler.HandleApproved(ctx, request, user)
		}

		if e != nil {
			mw.handleError(ctx, e)
			return
		}
		mw.handleSuccess(ctx, respFunc)
	}
}

func (mw *AuthorizeEndpointMiddleware) ApproveOrDenyHandlerFunc() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// no matter what happen, this is the last step. so clear saved request after done
		defer func() { _ = mw.clearAuthorizeRequest(ctx) }()

		// sanity checks
		request, client, user, e := mw.endpointSanityCheck(ctx)
		if e != nil {
			mw.handleError(ctx, e)
			return
		}
		logger.WithContext(ctx).Debug(fmt.Sprintf("AuthorizeRequest: %s", request))

		// parse approval params and check
		approval, remember := mw.parseApproval(ctx)
		if e := auth.ValidateApproval(ctx, approval, client, request.Scopes); e != nil {
			mw.handleError(ctx, e)
			return
		}
		request.Approved = true

		if remember {
			_ = mw.saveApprovedRequest(ctx, request)
		}

		// write response
		respFunc, e := mw.authorizeHandler.HandleApproved(ctx, request, user)
		if e != nil {
			mw.handleError(ctx, e)
			return
		}
		mw.handleSuccess(ctx, respFunc)
	}
}

func (mw *AuthorizeEndpointMiddleware) handleSuccess(c *gin.Context, v interface{}) {
	switch v.(type) {
	case auth.ResponseHandlerFunc:
		v.(auth.ResponseHandlerFunc)(c)
		c.Abort()
	default:
		c.JSON(200, v)
		c.Abort()
	}
}

func (mw *AuthorizeEndpointMiddleware) handleError(c *gin.Context, err error) {
	if !errors.Is(err, oauth2.ErrorTypeOAuth2) {
		err = oauth2.NewInvalidAuthorizeRequestError(err)
	}

	_ = mw.clearAuthorizeRequest(c)
	_ = c.Error(err)
	c.Abort()
}

func (mw *AuthorizeEndpointMiddleware) saveAuthorizeRequest(ctx *gin.Context, request *auth.AuthorizeRequest) error {
	s := session.Get(ctx)
	if s == nil {
		return oauth2.NewInternalError("failed to save authorize request for approval")
	}

	s.Set(sessionKeyAuthorizeRequest, request)
	if e := s.Save(); e != nil {
		return oauth2.NewInternalError("failed to save authorize request for approval", e)
	}
	return nil
}

func (mw *AuthorizeEndpointMiddleware) loadAuthorizeRequest(ctx *gin.Context) (*auth.AuthorizeRequest, error) {
	s := session.Get(ctx)
	if s == nil {
		return nil, oauth2.NewInternalError("failed to load authorize request for approval (no session)")
	}

	if request, ok := s.Get(sessionKeyAuthorizeRequest).(*auth.AuthorizeRequest); ok {
		return request.WithContext(context.Background()), nil
	}
	return nil, oauth2.NewInternalError("failed to load authorize request for approval")
}

func (mw *AuthorizeEndpointMiddleware) clearAuthorizeRequest(ctx *gin.Context) error {
	s := session.Get(ctx)
	if s == nil {
		return oauth2.NewInternalError("failed to clear authorize request for approval (no session)")
	}
	s.Delete(sessionKeyAuthorizeRequest)
	if e := s.Save(); e != nil {
		return oauth2.NewInternalError("failed to clear authorize request for approval", e)
	}
	return nil
}

func (mw *AuthorizeEndpointMiddleware) saveApprovedRequest(ctx *gin.Context, request *auth.AuthorizeRequest) error {
	if !request.Approved {
		return nil
	}
	s := session.Get(ctx)
	if s == nil {
		return oauth2.NewInternalError("failed to save approved request")
	}
	approved, _ := s.Get(sessionKeyApprovedRequests).([]*auth.AuthorizeRequest)
	approved = append(approved, request)
	s.Set(sessionKeyApprovedRequests, approved)
	if e := s.Save(); e != nil {
		return oauth2.NewInternalError("failed to save approved request", e)
	}
	return nil
}

func (mw *AuthorizeEndpointMiddleware) loadApprovedRequests(ctx *gin.Context) ([]*auth.AuthorizeRequest, error) {
	s := session.Get(ctx)
	if s == nil {
		return nil, oauth2.NewInternalError("failed to load approved requests")
	}
	approved, _ := s.Get(sessionKeyApprovedRequests).([]*auth.AuthorizeRequest)
	return approved, nil
}

func (mw *AuthorizeEndpointMiddleware) endpointSanityCheck(ctx *gin.Context) (
	*auth.AuthorizeRequest, oauth2.OAuth2Client, security.Authentication, error) {

	request, ok := ctx.Value(oauth2.CtxKeyValidatedAuthorizeRequest).(*auth.AuthorizeRequest)
	if !ok {
		return nil, nil, nil, oauth2.NewInternalError("authorize request not processed")
	}

	user := security.Get(ctx)
	if user.State() < security.StateAuthenticated {
		return nil, nil, nil, oauth2.NewInternalError("authorize endpoint is called without user authentication")
	}

	// retrieve client from context. It's should be populated by pre-auth MW
	client := auth.RetrieveAuthenticatedClient(ctx)
	if client == nil {
		return nil, nil, nil, oauth2.NewInternalError("client is not loaded")
	}

	return request, client, user, nil
}

func (mw *AuthorizeEndpointMiddleware) parseApproval(ctx *gin.Context) (approval map[string]bool, remember bool) {
	approved := false
	approval = make(map[string]bool)
	if v, ok := ctx.Request.PostForm[oauth2.ParameterUserApproval]; ok {
		approved, _ = strconv.ParseBool(v[len(v)-1])
	}
	if !approved {
		return
	}
	if v, ok := ctx.Request.PostForm[oauth2.ParameterRememberUserApproval]; ok {
		remember, _ = strconv.ParseBool(v[len(v)-1])
	}
	for k, v := range ctx.Request.PostForm {
		if !strings.HasPrefix(k, scopeParamPrefix) {
			continue
		}
		scope := strings.TrimPrefix(k, scopeParamPrefix)
		if len(v) > 0 {
			approval[scope], _ = strconv.ParseBool(v[len(v)-1])
		} else {
			approval[scope] = false
		}
	}
	return
}

func (mw *AuthorizeEndpointMiddleware) transferContextValues(src context.Context, dst context.Context) {
	mutable := utils.FindMutableContext(dst)
	listable, ok := src.(utils.ListableContext)
	if !ok || mutable == nil {
		return
	}
	for k, v := range listable.Values() {
		mutable.Set(k, v)
	}
}

func (mw *AuthorizeEndpointMiddleware) authRequestHasSavedApproval(ctx *gin.Context, request *auth.AuthorizeRequest) bool {
	approvedRequests, _ := mw.loadApprovedRequests(ctx)
	for _, r := range approvedRequests {
		if request.ClientId == r.ClientId &&
			request.RedirectUri == r.RedirectUri &&
			request.Scopes.Equals(r.Scopes) &&
			r.Approved {
			return true
		}
	}
	return false
}
