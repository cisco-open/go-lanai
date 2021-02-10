package authorize

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/auth"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/session"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"strconv"
	"strings"
)

const (
	sessionKeyAuthorizeRequest = "kAuthorizeRequest"
	scopeParamPrefix = "scope."
)
/***********************
	Authorize Endpoint
 ***********************/
type AuthorizeEndpointMiddleware struct {
	requestProcessor auth.AuthorizeRequestProcessor
	authorizeHanlder auth.AuthorizeHandler
	approveMatcher   web.RequestMatcher
	//TODO
}

type AuthorizeMWOptions func(*AuthorizeMWOption)

type AuthorizeMWOption struct {
	RequestProcessor auth.AuthorizeRequestProcessor
	AuthorizeHanlder auth.AuthorizeHandler
	ApprovalMatcher  web.RequestMatcher
	//TODO
}

func NewAuthorizeEndpointMiddleware(opts...AuthorizeMWOptions) *AuthorizeEndpointMiddleware {
	opt := AuthorizeMWOption{
		RequestProcessor: auth.NewCompositeAuthorizeRequestProcessor(),
		//TODO
	}
	for _, optFunc := range opts {
		if optFunc != nil {
			optFunc(&opt)
		}
	}
	return &AuthorizeEndpointMiddleware{
		requestProcessor: opt.RequestProcessor,
		authorizeHanlder: opt.AuthorizeHanlder,
		approveMatcher:   opt.ApprovalMatcher,
		// TODO
	}
}

func (mw *AuthorizeEndpointMiddleware) PreAuthenticateHandlerFunc() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// parse or load request
		var request *auth.AuthorizeRequest
		var err error
		switch approve, e := mw.approveMatcher.MatchesWithContext(ctx, ctx.Request); {
		case e == nil && approve:
			// approve or deny request
			if request, err = mw.loadAuthrozieRequest(ctx); err != nil {
				err = oauth2.NewInvalidAuthorizeRequestError("approval endpoint should be accessed via POST form", e)
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

func (mw *AuthorizeEndpointMiddleware) AuthroizeHandlerFunc() gin.HandlerFunc {
	return func(ctx *gin.Context) {
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
		if e != nil {
			// save request
			if e := mw.saveAuthrozieRequest(ctx, request); e != nil {
				mw.handleError(ctx, e)
				return
			}
			respFunc, e = mw.authorizeHanlder.HandleApprovalPage(ctx, request, user)
		} else {
			respFunc, e = mw.authorizeHanlder.HandleApproved(ctx, request, user)
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
		defer mw.clearAuthrozieRequest(ctx)

		// sanity checks
		request, client, user, e := mw.endpointSanityCheck(ctx)
		if e != nil {
			mw.handleError(ctx, e)
			return
		}
		logger.WithContext(ctx).Debug(fmt.Sprintf("AuthorizeRequest: %s", request))

		// parse approval params and check
		approval := mw.parseApproval(ctx)
		if e := auth.ValidateApproval(ctx, approval, client, request.Scopes); e != nil {
			mw.handleError(ctx, e)
			return
		}
		request.Approved = true

		// write response
		respFunc, e := mw.authorizeHanlder.HandleApproved(ctx, request, user)
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
		err = oauth2.NewInvalidAuthorizeRequestError(err.Error(), err)
	}

	mw.clearAuthrozieRequest(c)
	_ = c.Error(err)
	c.Abort()
}

func (mw *AuthorizeEndpointMiddleware) saveAuthrozieRequest(ctx *gin.Context, request *auth.AuthorizeRequest) (error) {
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

func (mw *AuthorizeEndpointMiddleware) loadAuthrozieRequest(ctx *gin.Context) (*auth.AuthorizeRequest, error) {
	s := session.Get(ctx)
	if s == nil {
		return nil, oauth2.NewInternalError("failed to load authorize request for approval (no session)")
	}

	if request, ok := s.Get(sessionKeyAuthorizeRequest).(*auth.AuthorizeRequest); ok {
		return request.WithContext(utils.NewMutableContext()), nil
	}
	return nil, oauth2.NewInternalError("failed to load authorize request for approval")
}

func (mw *AuthorizeEndpointMiddleware) clearAuthrozieRequest(ctx *gin.Context) error {
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

func (mw *AuthorizeEndpointMiddleware) parseApproval(ctx *gin.Context) (approval map[string]bool) {
	approval = make(map[string]bool)
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

func (mw *AuthorizeEndpointMiddleware) transferContextValues(src context.Context, dst *gin.Context) {
	listable, ok := src.(utils.ListableContext)
	if !ok {
		return
	}

	for k, v := range listable.Values() {
		key := ""
		switch k.(type) {
		case string:
			key = k.(string)
		case fmt.Stringer:
			key = k.(fmt.Stringer).String()
		default:
			continue
		}
		dst.Set(key, v)
	}
}