package health

import (
	"context"
	"github.com/gin-gonic/gin"
)

type PingIndicator struct {

}

func (b PingIndicator) Name() string {
	return "ping"
}

func (b PingIndicator) Health(ctx context.Context, options Options) Health {
	// very basic check: if the given context is *gin.Context, it means the health check is invoked via web endpoint.
	// therefore the web framework is still working
	if _, ok := ctx.(*gin.Context); ok {
		return NewDetailedHealth(StatusUp, "", map[string]interface{}{
			"Reason": "context is web context",
		})
	}
	return NewDetailedHealth(StatusUnkown, "", map[string]interface{}{
		"Reason": "context is not web context"})
}

