package access

import (
	"github.com/gin-gonic/gin"
)

//goland:noinspection GoNameStartsWithPackageName
type AccessControlMiddleware struct {
	decisionMakers []DecisionMakerFunc
}

func NewAccessControlMiddleware(decisionMakers...DecisionMakerFunc) *AccessControlMiddleware {
	return &AccessControlMiddleware{decisionMakers: decisionMakers}
}

func (ac *AccessControlMiddleware) ACHandlerFunc() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var err error
		for _, decisionMaker := range ac.decisionMakers {
			var handled bool
			handled, err = decisionMaker(ctx, ctx.Request)
			if handled {
				break
			}
		}

		if err != nil {
			// access denied
			ac.handleError(ctx, err)
		} else {
			ctx.Next()
		}
	}
}

func (ac *AccessControlMiddleware) handleError(c *gin.Context, err error) {
	// We add the error and let the error handling middleware to render it
	_ = c.Error(err)
	c.Abort()
}
