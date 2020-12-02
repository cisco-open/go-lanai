package web

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/matcher"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	httptransport "github.com/go-kit/kit/transport/http"
	"go.uber.org/fx"
	"net/http"
	"path"
	"reflect"
	"sort"
	"time"
)

const (
	kGinContextKey = "GinContext"
	DefaultGroup = "/"
)

var (
	bindingValidator binding.StructValidator
)

type Registrar struct {
	engine            *gin.Engine
	router            gin.IRouter
	properties        ServerProperties
	options           []httptransport.ServerOption // options go-kit middleware options
	validator         binding.StructValidator
	middlewares       []MiddlewareMapping // middlewares gin-gonic middleware providers
}

// TODO support customizers
func NewRegistrar(g *gin.Engine, properties ServerProperties) *Registrar {

	var contextPath = path.Clean("/" + properties.ContextPath)
	registrar := &Registrar{
		engine:     g,
		router: 	g.Group(contextPath),
		properties: properties,
		options: []httptransport.ServerOption{
			httptransport.ServerBefore(ginContextExtractor),
		},
		validator: binding.Validator,
	}

	// add some global middlewares
	_ = registrar.addGlobalMiddleware("pre-process", HighestMiddlewareOrder, registrar.preProcessMiddleware)
	return registrar
}

// initialize should be called during application startup, last change to change configurations, load templates, etc
func (r *Registrar) initialize() (err error) {
	// TODO support customizers
	r.engine.LoadHTMLGlob("web/template/*")
	// we disable auto-validation. We will invoke our own validation manually.
	// Also we need to make the validator available globally for any request decoder to access.
	// The alternative approach is to put the validator into each gin.Context
	binding.Validator = nil
	bindingValidator = r.validator
	return
}

// addGlobalMiddleware add middleware to all mapping
func (r *Registrar) addGlobalMiddleware(name string, order int, handlerFunc gin.HandlerFunc) error {
	mapping := NewMiddlewareMapping(name, order, matcher.Any(), handlerFunc)
	return r.Register(mapping)
}

// Run configure and start gin engine
func (r *Registrar) Run() (err error) {
	if err = r.initialize(); err != nil {
		return
	}

	var addr = fmt.Sprintf(":%v", r.properties.Port)
	s := &http.Server{
		Addr:           addr,
		Handler:        r.engine,
		ReadTimeout:    60 * time.Second,
		WriteTimeout:   60 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	go s.ListenAndServe()
	return
}

// Register is the entry point to register Controller, Mapping and other web related objects
// supported items type are:
// 	- Controller
//  - EndpointMapping
//  - StaticMapping
//  - TemplateMapping
//  - MiddlewareMapping
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
	case MiddlewareMapping:
		err = r.registerMiddlewareMapping(i.(MiddlewareMapping))
	case GenericMapping:
		err = r.registerGenericMapping(i.(GenericMapping))
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
	endpoints := c.Mappings()
	for _, m := range endpoints {
		if err = r.register(m); err != nil {
			err = fmt.Errorf("invalid endpoint mapping in Controller [%T]: %v", c, err.Error())
			break
		}
	}
	return
}

func (r *Registrar) registerMvcMapping(m MvcMapping) error {
	options := r.options
	if m.ErrorEncoder() != nil {
		options = append(r.options, httptransport.ServerErrorEncoder(m.ErrorEncoder()))
	}

	s := httptransport.NewServer(
		m.Endpoint(),
		m.DecodeRequestFunc(),
		m.EncodeResponseFunc(),
		options...,
	)

	handlerFunc := MakeGinHandlerFunc(s)
	middlewares, err := r.findMiddlewares(DefaultGroup, m.Path(), m.Method())
	r.router.Group(DefaultGroup).
		Use(middlewares...).
		Handle(m.Method(), m.Path(), handlerFunc)
	return err
}

func (r *Registrar) registerStaticMapping(m StaticMapping) error {
	// TODO handle suffix rewrite, e.g. /path/to/swagger -> /path/to/swagger.html
	middlewares, err := r.findMiddlewares(DefaultGroup, m.Path(), http.MethodGet, http.MethodHead)
	r.router.Group(DefaultGroup).
		Use(middlewares...).
		Static(m.Path(), m.StaticRoot())
	return err
}

func (r *Registrar) registerMiddlewareMapping(m MiddlewareMapping) error {
	r.middlewares = append(r.middlewares, m)
	return nil
}

func (r *Registrar) registerGenericMapping(m GenericMapping) error {
	middlewares, err := r.findMiddlewares(DefaultGroup, m.Path(), m.Method())
	r.router.Group(DefaultGroup).
		Use(middlewares...).
		Handle(m.Method(), m.Path(), m.HandlerFunc())
	return err
}

func (r *Registrar) findMiddlewares(group, relativePath string, methods...string) (gin.HandlersChain, error) {
	var handlers = make([]gin.HandlerFunc, len(r.middlewares))
	sort.SliceStable(r.middlewares, func(i,j int) bool { return r.middlewares[i].Order() < r.middlewares[j].Order()})
	var i = 0
	for _,mw := range r.middlewares {
		switch match, err := r.routeMatches(mw.Matcher(), group, relativePath, methods...); {
		case err != nil:
			return []gin.HandlerFunc{}, err
		case match:
			handlers[i] = mw.HandlerFunc()
			i++
		}
	}
	return handlers[:i], nil
}

func (r *Registrar) routeMatches(matcher RouteMatcher, group, relativePath string, methods...string) (bool, error) {
	switch {
	case len(methods) == 0:
		return false, fmt.Errorf("unable to register middleware: method is missing for %s", relativePath)
	case matcher == nil:
		return true, nil // no matcher, any value is a match
	}

	// match if any given method matches
	for _,m := range methods {
		ret, err := matcher.Matches(Route{Group: group, Path: relativePath, Method: m})
		if ret || err != nil {
			return ret, err
		}
	}
	return false, nil
}

/*******************************
	some global middlewares
********************************/
func (r *Registrar) preProcessMiddleware(c *gin.Context) {
	c.Set(ContextKeyContextPath, r.properties.ContextPath)
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






