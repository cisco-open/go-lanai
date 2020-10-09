package web

import (
	"context"
	"github.com/gin-gonic/gin"
	httptransport "github.com/go-kit/kit/transport/http"
	"net/http"
)

const (
	kGinContextKey = "GinContext"
)

type Registrar struct {
	engine *gin.Engine
	options []httptransport.ServerOption
}

// TODO support customizers
func NewRegistrar(g *gin.Engine) *Registrar {
	return &Registrar{
		engine: g,
		options: []httptransport.ServerOption{
			httptransport.ServerBefore(ginContextExtractor),
		},
	}
}

func (r *Registrar) Register(mapping interface{}) {
	// TODO better error handling
	switch mapping.(type) {
	case *Mapping:
		r.registerByMapping(mapping.(*Mapping))
	}
}

func (r *Registrar) registerByMapping(m *Mapping) {
	s := httptransport.NewServer(
		m.Endpoint,
		m.DecodeRequestFunc,
		m.EncodeResponseFunc,
		r.options...,
	)

	handlerFunc := MakeGinHandlerFunc(s)
	if m.MappingFunc != nil {
		m.MappingFunc(r.engine, handlerFunc)
	} else {
		r.engine.Handle(m.Method, m.Path, handlerFunc)
	}
}

/**************************
	first class functions
***************************/

func MakeGinHandlerFunc(s *httptransport.Server) gin.HandlerFunc {
	return func(c *gin.Context) {
		reqCtx := context.WithValue(c.Request.Context(), kGinContextKey, c)
		c.Request = c.Request.WithContext(reqCtx)
		s.ServeHTTP(c.Writer, c.Request)
	}
}

func ginContextExtractor(ctx context.Context, r *http.Request) (ret context.Context) {
	if ret = r.Context().Value(kGinContextKey).(context.Context); ret == nil {
		return ctx
	}
	return
}




