package web

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

type RequestPreProcessorName string

type RequestPreProcessor interface {
	Process(r *http.Request) error
	Name() RequestPreProcessorName
}

type Engine struct {
	*gin.Engine
	requestPreProcessor []RequestPreProcessor
}

func (e *Engine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for _, p := range e.requestPreProcessor {
		err := p.Process(r)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = fmt.Fprint(w, "Internal error with request cache")
			return
		}
	}
	e.Engine.ServeHTTP(w, r)
}

func (e *Engine) addRequestPreProcessor(p RequestPreProcessor) {
	e.requestPreProcessor = append(e.requestPreProcessor, p)
}

func NewEngine() *Engine {
	if bootstrap.DebugEnabled() {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	eng := &Engine{
		Engine: gin.New(),
	}
	//eng.ContextWithFallback = true
	return eng
}