package session

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"fmt"
	"github.com/gin-gonic/gin"
	"math/rand"
	"time"
)

//TODO: remove
func SessionDebugHandlerFunc() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		auth := security.Get(ctx)
		if auth.Authenticated() {
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
