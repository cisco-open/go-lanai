package session

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"github.com/gin-gonic/gin"
	"math/rand"
	"net/http"
	"time"
)

//TODO: remove
func SessionDebugHandlerFunc() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		auth := security.Get(ctx)
		if auth.State() > security.StateAnonymous {
			logger.Debugf("Already authenticated as %T", auth)
		}

		session := Get(ctx)
		if session.Get("TEST") == nil {

			session.Set("TEST", RandomString(10240))
		} else {
			logger.Debugf("Have Session Token %s", "TEST")
		}

		ctx.Next()
	}
}

const charset = "abcdefghijklmnopqrstuvwxyz_ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func RandomString(length int) string {
	var seededRand *rand.Rand = rand.New(
		rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}


type DebugAuthSuccessHandler struct {}

func (h *DebugAuthSuccessHandler) HandleAuthenticationSuccess(
	_ context.Context, _ *http.Request, _ http.ResponseWriter, from, to security.Authentication) {
	logger.Debugf("session knows auth succeeded: from [%v] to [%v]", from, to)
}

type DebugAuthErrorHandler struct {}

func (h *DebugAuthErrorHandler) HandleAuthenticationError(_ context.Context, _ *http.Request, _ http.ResponseWriter, err error) {
	logger.Debugf("session knows auth failed with %v", err.Error())
}
