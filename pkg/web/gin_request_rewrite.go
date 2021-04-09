package web

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

type ginRequestRewriter struct {
	engine *gin.Engine
}

func newGinRequestRewriter(engine *gin.Engine) RequestRewriter {
	return &ginRequestRewriter{
		engine: engine,
	}
}

// Caution, you could loop yourself to death
func (rw ginRequestRewriter) HandleRewrite(r *http.Request) error {
	gc := GinContext(r.Context())
	if gc == nil {
		return fmt.Errorf("the request is not linked to a gin Context. Please make sure this is the right RequestRewriter to use")
	}

	rw.engine.HandleContext(gc)
	return nil
}
