package basicauth

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/passwd"
	"encoding/base64"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"strings"
)

//goland:noinspection GoNameStartsWithPackageName
type BasicAuthMiddleware struct {
	authenticator security.Authenticator
	successHandler security.AuthenticationSuccessHandler
}

func NewBasicAuthMiddleware(authenticator security.Authenticator, successHandler security.AuthenticationSuccessHandler) *BasicAuthMiddleware {
	return &BasicAuthMiddleware{
		authenticator:  authenticator,
		successHandler: successHandler,
	}
}

func (basic *BasicAuthMiddleware) HandlerFunc() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		header := ctx.GetHeader("Authorization")
		if header == "" {
			// Authorization header not available, bail
			basic.handleSuccess(ctx, nil)
			return
		}
		encoded := strings.TrimLeft(header, "Basic ")
		decoded, err := base64.StdEncoding.DecodeString(encoded)
		if err != nil {
			basic.handleError(ctx, security.NewBadCredentialsError("invalid Authorization header"))
			return
		}

		pair := strings.SplitN(string(decoded), ":", 2)
		if len(pair) < 2 {
			basic.handleError(ctx, security.NewBadCredentialsError("invalid Authorization header"))
			return
		}

		currentAuth, ok := security.Get(ctx).(passwd.UsernamePasswordAuthentication)
		if ok && currentAuth.Authenticated() && passwd.IsSamePrincipal(pair[0], currentAuth) {
			// already authenticated
			basic.handleSuccess(ctx, nil)
			return
		}

		candidate := passwd.UsernamePasswordPair{
			Username: pair[0],
			Password: pair[1],
		}
		// Search auth in the slice of allowed credentials
		auth, err := basic.authenticator.Authenticate(&candidate)
		if err != nil {
			basic.handleError(ctx, err)
			return
		}

		// The auth credentials was found, set auth's id to key AuthUserKey in this context, the auth's id can be read later using
		// c.MustGet(gin.AuthUserKey).
		basic.handleSuccess(ctx, auth)
	}
}

func (basic *BasicAuthMiddleware) handleSuccess(c *gin.Context, new security.Authentication) {
	if new != nil {
		c.Set(gin.AuthUserKey, new.Principal())
		c.Set(security.ContextKeySecurity, new)
		basic.successHandler.HandleAuthenticationSuccess(c, c.Request, c.Writer, new)
	}
	c.Next()
}

func (basic *BasicAuthMiddleware) handleError(c *gin.Context, err error) {
	_ = c.Error(err)
	c.Abort()
}

//goland:noinspection GoNameStartsWithPackageName
type BasicAuthEntryPoint struct {}

func NewBasicAuthEntryPoint() *BasicAuthEntryPoint {
	return &BasicAuthEntryPoint{}
}

func (h *BasicAuthEntryPoint) Commence(_ context.Context, _ *http.Request, rw http.ResponseWriter, err error) {
	writeBasicAuthChallenge(rw, err)
}

//goland:noinspection GoNameStartsWithPackageName
type BasicAuthErrorHandler struct {}

func NewBasicAuthErrorHandler() *BasicAuthErrorHandler {
	return &BasicAuthErrorHandler{}
}

func (h *BasicAuthErrorHandler) HandleAuthenticationError(_ context.Context, _ *http.Request, rw http.ResponseWriter, err error) {
	writeBasicAuthChallenge(rw, err)
}

func writeBasicAuthChallenge(rw http.ResponseWriter, err error) {
	realm := "Basic realm=" + strconv.Quote("Authorization Required")
	rw.Header().Set("WWW-Authenticate", realm)
	rw.WriteHeader(http.StatusUnauthorized)
	_,_ = rw.Write([]byte(err.Error()))
}