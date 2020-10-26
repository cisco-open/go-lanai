package basic_auth

import (
	"cto-github.cisco.com/livdu/jupiter/pkg/security"
	"cto-github.cisco.com/livdu/jupiter/pkg/web"
	"encoding/base64"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"strings"
)

type Authenticator interface {
	AuthHandler() gin.HandlerFunc
}

type BasicAuth struct {
	store security.AccountStore
}

func NewBasicAuth(store security.AccountStore) Authenticator {
	return &BasicAuth{store}
}

func (auth *BasicAuth) ConditionFunc() web.ConditionalMiddlewareFunc {
	return func(r *http.Request) bool {
		return true
	}
}
func (auth *BasicAuth) HandlerFunc() gin.HandlerFunc {
	return auth.AuthHandler()
}

func (auth *BasicAuth) AuthHandler() gin.HandlerFunc {
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
		// Search user in the slice of allowed credentials
		user, err := auth.store.LoadUserByUsername(pair[0])
		if err != nil {
			ctx.Header("WWW-Authenticate", realm)
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// Check password
		if pair[0] != user.Username() || pair[1] != user.Password() {
			ctx.Header("WWW-Authenticate", realm)
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// The user credentials was found, set user's id to key AuthUserKey in this context, the user's id can be read later using
		// c.MustGet(gin.AuthUserKey).
		ctx.Set(gin.AuthUserKey, user)
		ctx.Next()
	}
}
