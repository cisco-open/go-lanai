package basicauth

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/passwd"
	"encoding/base64"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"strings"
)

type BasicAuthMiddleware struct {
	authenticator security.Authenticator
}

func NewBasicAuthMiddleware(auth security.Authenticator) *BasicAuthMiddleware {
	return &BasicAuthMiddleware{auth}
}

func (basic *BasicAuthMiddleware) HandlerFunc() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		currentAuth, ok := security.Get(ctx).(passwd.UsernamePasswordAuthentication)
		if ok && currentAuth.Authenticated() {
			// already authenticated
			basic.handleSuccess(ctx, nil)
			return
		}

		header := ctx.GetHeader("Authorization")
		encoded := strings.TrimLeft(header, "Basic ")
		decoded, err := base64.StdEncoding.DecodeString(encoded)
		if err != nil {
			basic.handleError(ctx, err)
			return
		}

		pair := strings.SplitN(string(decoded), ":", 2)
		if len(pair) < 2 {
			basic.handleError(ctx, err)
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
	// TODO delegate to authentication handler to support session fixation and more
	if new != nil {
		c.Set(gin.AuthUserKey, new.Principal())
		c.Set(security.ContextKeySecurity, new)
	}
	c.Next()
}

func (basic *BasicAuthMiddleware) handleError(c *gin.Context, err error) {
	realm := "Basic realm=" + strconv.Quote("Authorization Required")
	c.Header("WWW-Authenticate", realm)
	_ = c.AbortWithError(http.StatusUnauthorized, err)
}
