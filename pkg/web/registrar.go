package web

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	httptransport "github.com/go-kit/kit/transport/http"
	"go.uber.org/fx"
	"net/http"
	"reflect"
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

// initialize should be called during application startup, last change to change configurations, load templates, etc
func (r *Registrar) initialize() (err error) {
	// TODO support customizers
	r.engine.LoadHTMLGlob("web/template/*")
	return
}

// Register is the entry point to register Controller, Mapping and other web related objects
// supported items type are:
// 	- Controller
//  - EndpointMapping
//  - StaticMapping
//  - TemplateMapping
//  - struct that contains exported Controller fields
func (r *Registrar) Register(items...interface{}) (err error) {
	for _, i := range items {
		if err = r.register(i); err != nil {
			break
		}
	}
	return
}

// RegisterWithLifecycle is a convenient function to schedule item registration in FX lifecycle
func (r *Registrar) RegisterWithLifecycle(lc fx.Lifecycle, items...interface{}) {
	lc.Append(fx.Hook{
		OnStart: func(context.Context) (err error) {
			return r.Register(items...)
		},
	})
}

func (r *Registrar) register(i interface{}) (err error) {
	switch i.(type) {
	case Controller:
		err = r.registerController(i.(Controller))
	case MvcMapping:
		err = r.registerMvcMapping(i.(MvcMapping))
	case StaticMapping:
		err = r.registerStaticMapping(i.(StaticMapping))
	default:
		err = r.registerUnknownType(i)
	}
	return
}

func (r *Registrar) registerUnknownType(i interface{}) (err error) {
	v := reflect.ValueOf(i)

	// get struct value
	if v.Kind() == reflect.Ptr && v.Elem().Kind() == reflect.Struct {
		v = v.Elem()
	} else if v.Kind() != reflect.Struct {
		return errors.New(fmt.Sprintf("unsupported type [%T]", i))
	}

	// go through fields and register
	for idx := 0; idx < v.NumField(); idx++ {
		// only care controller fields
		c := v.Field(idx).Interface()
		if _,ok := c.(Controller); !ok {
			continue
		}

		err = r.register(c)
		if err != nil {
			return err
		}
	}
	return
}

func (r *Registrar) registerController(c Controller) (err error) {
	endpoints := c.Endpoints()
	for _, m := range endpoints {
		if err = r.register(m); err != nil {
			err = fmt.Errorf("invalid endpoint mapping in Controller [%T]: %v", c, err.Error())
			break
		}
	}
	return
}

func (r *Registrar) registerMvcMapping(m EndpointMapping) error {
	s := httptransport.NewServer(
		m.Endpoint(),
		m.DecodeRequestFunc(),
		m.EncodeResponseFunc(),
		r.options...,
	)

	handlerFunc := MakeGinHandlerFunc(s)
	r.engine.Handle(m.Method(), m.Path(), handlerFunc)
	return nil
}

func (r *Registrar) registerStaticMapping(m StaticMapping) error {
	// TODO handle suffix rewrite, e.g. /path/to/swagger -> /path/to/swagger.html
	r.engine.Static(m.Path(), m.StaticRoot())
	return nil
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




