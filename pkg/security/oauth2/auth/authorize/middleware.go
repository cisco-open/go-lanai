package authorize

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/auth"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
)

/***********************
	Token Endpoint
 ***********************/
type AuhtorizeEndpointMiddleware struct {
	requestProcessor auth.AuthorizeRequestProcessor
	//TODO
}

type AuthorizeMWOptions func(*AuthorizeMWOption)

type AuthorizeMWOption struct {
	RequestProcessor auth.AuthorizeRequestProcessor
	//TODO
}

func NewTokenEndpointMiddleware(opts...AuthorizeMWOptions) *AuhtorizeEndpointMiddleware {
	opt := AuthorizeMWOption{
		RequestProcessor: auth.NewCompositeAuthorizeRequestProcessor(),
		//TODO
	}
	for _, optFunc := range opts {
		if optFunc != nil {
			optFunc(&opt)
		}
	}
	return &AuhtorizeEndpointMiddleware{
		requestProcessor: opt.RequestProcessor,
		// TODO
	}
}

func (mw *AuhtorizeEndpointMiddleware) PreAuthenticateHandlerFunc() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// parse request
		request, e := auth.ParseAuthorizeRequest(ctx.Request)
		if e != nil {
			mw.handleError(ctx, oauth2.NewInvalidAuthorizeRequestError("invalid authorize request", e))
			return
		}
		ctx.Set(oauth2.CtxKeyReceivedAuthorizeRequest, request)

		// validate and process, regardless the result, we might want to transfer some context from request to current context
		processed, e := mw.requestProcessor.Process(ctx, request)
		mw.transferContextValues(request.Context(), ctx)
		if e != nil {
			mw.handleError(ctx, e)
			return
		}

		// everything is ok, set it to context for later usage
		ctx.Set(oauth2.CtxKeyValidatedAuthorizeRequest, processed)
	}
}

func (mw *AuhtorizeEndpointMiddleware) AuthroizeHandlerFunc() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		request, ok := ctx.Value(oauth2.CtxKeyValidatedAuthorizeRequest).(*auth.AuthorizeRequest)
		if !ok {
			mw.handleError(ctx, oauth2.NewInternalError("authorize request not processed"))
		}

		// TODO do something with request
		logger.WithContext(ctx).Info("Recieved authorize request: %s", request)
		ctx.JSON(200, map[string]string {
			"TODO": "authorize endpoint is not implemented yet",
		})
	}
}

// TODO
func (mw *AuhtorizeEndpointMiddleware) handleSuccess(c *gin.Context, v interface{}) {
	c.JSON(200, v)
	c.Abort()
}

// TODO
func (mw *AuhtorizeEndpointMiddleware) handleError(c *gin.Context, err error) {
	if !errors.Is(err, oauth2.ErrorTypeOAuth2) {
		err = oauth2.NewInvalidAuthorizeRequestError(err.Error(), err)
	}

	security.Clear(c)
	_ = c.Error(err)
	c.Abort()
}

func (mw *AuhtorizeEndpointMiddleware) transferContextValues(src context.Context, dst *gin.Context) {
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