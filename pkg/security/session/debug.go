package session

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"fmt"
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
			fmt.Printf("Already authenticated as %T\n", auth)
		}

		session := Get(ctx)
		if session.Get("TEST") == nil {

			session.Set("TEST", RandomString(10240))
		} else {
			fmt.Printf("Have Session Value %s\n", "TEST")
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
	fmt.Printf("[DEBUG] session knows auth succeeded: from [%v] to [%v] \n", from, to)
}

type DebugAuthErrorHandler struct {}

func (h *DebugAuthErrorHandler) HandleAuthenticationError(_ context.Context, _ *http.Request, _ http.ResponseWriter, err error) {
	fmt.Printf("[DEBUG] session knows auth failed with %v \n", err.Error())
}
