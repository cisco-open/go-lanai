package health

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
)

type PingIndicator struct {

}

func (b PingIndicator) Name() string {
	return "ping"
}

func (b PingIndicator) Health(ctx context.Context, options Options) Health {
	// very basic check: if the given context is *gin.Context, it means the health check is invoked via web endpoint.
	// therefore the web framework is still working
	if g := web.GinContext(ctx); g != nil {
		return NewDetailedHealth(StatusUp, "", nil)
	}
	return NewDetailedHealth(StatusUnknown, "", nil)
}

