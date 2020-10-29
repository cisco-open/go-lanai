package basicauth

import (
	"cto-github.cisco.com/livdu/jupiter/pkg/security"
	"cto-github.cisco.com/livdu/jupiter/pkg/security/passwd"
	"encoding/base64"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"strings"
)

type BasicAuthMiddleware struct {
	authenticator security.Authenticator
}

func NewBasicAuthMiddleware(store security.Authenticator) *BasicAuthMiddleware {
	return &BasicAuthMiddleware{store}
}

//func (basic *BasicAuthMiddleware) ConditionFunc() web.ConditionalMiddlewareFunc {
//	return func(r *http.Request) bool {
//		return true
//	}
//}

func (basic *BasicAuthMiddleware) HandlerFunc() gin.HandlerFunc {
	realm := "Basic realm=" + strconv.Quote("Authorization Required")
	return func(ctx *gin.Context) {
		header := ctx.GetHeader("Authorization")
		encoded := strings.TrimLeft(header, "Basic ")
		decoded, err := base64.StdEncoding.DecodeString(encoded)
		if err != nil {
			ctx.Header("WWW-Authenticate", realm)
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		pair := strings.SplitN(string(decoded), ":", 2)
		if len(pair) < 2 {
			ctx.Header("WWW-Authenticate", realm)
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		candidate := passwd.UsernamePasswordPair{
			Username: pair[0],
			Password: pair[1],
		}
		// Search auth in the slice of allowed credentials
		auth, err := basic.authenticator.Authenticate(&candidate)
		if err != nil {
			ctx.Header("WWW-Authenticate", realm)
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// The auth credentials was found, set auth's id to key AuthUserKey in this context, the auth's id can be read later using
		// c.MustGet(gin.AuthUserKey).
		ctx.Set(gin.AuthUserKey, auth)
		ctx.Next()
	}
}
