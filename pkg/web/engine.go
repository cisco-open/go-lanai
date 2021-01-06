package web

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

type RequestPreProcessorProvider interface {
	GetPreProcessor() RequestPreProcessor
}

type RequestPreProcessor interface {
	Process(r *http.Request) error
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
	e := &Engine{
		Engine: gin.Default(),
	}
	return e
}