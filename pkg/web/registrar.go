package web

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	. "cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/matcher"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/reflectutils"
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
	"strings"
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
	engine         *Engine
	router         gin.IRouter
	properties     ServerProperties
	options        []httptransport.ServerOption // options go-kit middleware options
	validator      binding.StructValidator
	middlewares    []MiddlewareMapping                   // middlewares gin-gonic middleware providers
	routedMappings map[string]map[string][]RoutedMapping // routedMappings MvcMappings + SimpleMappings
	staticMappings []StaticMapping                       // staticMappings all static mappings
	customizers    []Customizer
	initialized    bool
}

func NewRegistrar(g *Engine, properties ServerProperties) *Registrar {

	var contextPath = path.Clean("/" + properties.ContextPath)
	registrar := &Registrar{
		engine:     g,
		router:     g.Group(contextPath),
		properties: properties,
		options: []httptransport.ServerOption{
			httptransport.ServerBefore(ginContextExtractor),
		},
		validator:      binding.Validator,
		routedMappings: map[string]map[string][]RoutedMapping{},
	}

	registrar.AddGlobalMiddlewares(gin.Recovery())
	return registrar
}

// initialize should be called during application startup, last change to change configurations, load templates, etc
func (r *Registrar) initialize() (err error) {
	if r.initialized {
		return fmt.Errorf("attempting to initialize web engine multiple times")
	}

	// apply customizers before install mappings
	if err = r.applyCustomizers(); err != nil {
		return
	}

	// we disable auto-validation. We will invoke our own validation manually.
	// Also we need to make the validator available globally for any request decoder to access.
	// The alternative approach is to put the validator into each gin.Context
	binding.Validator = nil
	bindingValidator = r.validator

	// load templates
	r.engine.LoadHTMLGlob("web/template/*")

	// add some common middlewares
	mappings := []interface{}{
		NewMiddlewareMapping("pre-process", HighestMiddlewareOrder, Any(), r.preProcessMiddleware),
	}
	if err = r.Register(mappings...); err != nil {
		return
	}

	// before starting to register mappings, we want global MW to take effect on our main group
	var contextPath = path.Clean("/" + r.properties.ContextPath)
	r.router = r.engine.Group(contextPath)

	// register routedMappings to gin engine
	if err = r.installMappings(); err != nil {
		return
	}

	r.initialized = true
	return
}

// AddGlobalMiddleware add middleware to all mapping
func (r *Registrar) AddGlobalMiddlewares(handlerFuncs ...gin.HandlerFunc) error {
	r.engine.Use(handlerFuncs...)
	return nil
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
// 	- Customizer
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
	if r.initialized {
		return errors.New("attempting to register mappings/middlewares/pre-processors after web engine initialization")
	}

	switch i.(type) {
	case Controller:
		err = r.registerController(i.(Controller))
	case MvcMapping:
		err = r.registerMvcMapping(i.(MvcMapping))
	case StaticMapping:
		err = r.registerStaticMapping(i.(StaticMapping))
	case MiddlewareMapping:
		err = r.registerMiddlewareMapping(i.(MiddlewareMapping))
	case SimpleMapping:
		err = r.registerSimpleMapping(i.(SimpleMapping))
	case RequestPreProcessor:
		err = r.registerRequestPreProcessor(i.(RequestPreProcessor))
	case Customizer:
		err = r.registerWebCustomizer(i.(Customizer))
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
		if f := v.Type().Field(idx); !reflectutils.IsExportedField(f) {
			// unexported field
			continue
		}
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

func (r *Registrar) registerRoutedMapping(m RoutedMapping) error {
	method := strings.ToUpper(m.Method())
	path := NormalizedPath(m.Path())

	paths, ok := r.routedMappings[method]
	if !ok {
		paths = map[string][]RoutedMapping{}
		r.routedMappings[method] = paths
	}

	if mappings, ok := paths[path]; !ok {
		paths[path] = []RoutedMapping{m}
	} else {
		paths[path] = append(mappings, m)
	}
	return nil
}

func (r *Registrar) registerSimpleMapping(m SimpleMapping) error {
	return r.registerRoutedMapping(m)
}

func (r *Registrar) registerMvcMapping(m MvcMapping) error {
	return r.registerRoutedMapping(m)
}

func (r *Registrar) registerStaticMapping(m StaticMapping) error {
	r.staticMappings = append(r.staticMappings, m)
	return nil
}

func (r *Registrar) registerMiddlewareMapping(m MiddlewareMapping) error {
	r.middlewares = append(r.middlewares, m)
	return nil
}

func (r *Registrar) registerRequestPreProcessor(p RequestPreProcessor) error {
	r.engine.addRequestPreProcessor(p)
	return nil
}

func (r *Registrar) registerWebCustomizer(c Customizer) error {
	if r.initialized {
		return fmt.Errorf("cannot register web configurer after web engine have initialized")
	}
	r.customizers = append(r.customizers, c)
	return nil
}

func (r *Registrar) applyCustomizers() error {
	if r.customizers == nil {
		return nil
	}
	for _, c := range r.customizers {
		if e := c.Customize(r); e != nil {
			return e
		}
	}
	return nil
}

func (r *Registrar) installMappings() error {
	// register routedMappings
	for method, paths := range r.routedMappings {
		for _, mappings := range paths {
			// all routedMappings with condition registered first
			sort.SliceStable(mappings, func(i,j int) bool {
				return mappings[i].Condition() != nil && mappings[j].Condition() == nil
			})

			if e := r.installRoutedMappings(method, mappings); e != nil {
				return e
			}
		}
	}

	// register static mappings
	for _,m := range r.staticMappings {
		if e := r.installStaticMapping(m); e != nil {
			return e
		}
	}
	return nil
}

func (r *Registrar) installStaticMapping(m StaticMapping) error {
	// TODO handle suffix rewrite, e.g. /path/to/swagger -> /path/to/swagger.html
	middlewares, err := r.findMiddlewares(DefaultGroup, m.Path(), http.MethodGet, http.MethodHead)
	r.router.Group(DefaultGroup).
		Use(middlewares...).
		Static(m.Path(), m.StaticRoot())
	return err
}

func (r *Registrar) installRoutedMappings(method string, mappings []RoutedMapping) error {
	if len(mappings) == 0 {
		return nil
	}

	handlerFuncs := make([]gin.HandlerFunc, len(mappings))
	path := mappings[0].Path()
	unconditionalFound := false
	for i, m := range mappings {
		// validate method and path with best efforts
		switch {
		case path != m.Path():
			return fmt.Errorf("attempt to register multiple RoutedMappings with inconsist path parameters: " +
				"expected [%s %s] but got [%s %s]", method, path, m.Method(), m.Path())
		case m.Condition() == nil && unconditionalFound:
			return fmt.Errorf("attempt to register multiple unconditional RoutedMappings on same path and method: [%s %s]", m.Method(), m.Path())
		case m.Condition() == nil:
			unconditionalFound = true
		}

		// create hander funcs
		switch m.(type) {
		case MvcMapping:
			handlerFuncs[i] = r.makeHandlerFuncFromMvcMapping(m.(MvcMapping))
		case SimpleMapping:
			handlerFuncs[i] = MakeConditionalHandlerFunc(m.(SimpleMapping).HandlerFunc(), m.Condition())
		}
	}

	middlewares, err := r.findMiddlewares(DefaultGroup, path, method)
	if method == MethodAny {
		r.router.Group(DefaultGroup).
			Use(middlewares...).
			Any(path, handlerFuncs...)
	} else {
		r.router.Group(DefaultGroup).
			Use(middlewares...).
			Handle(method, path, handlerFuncs...)
	}

	return err
}

func (r *Registrar) makeHandlerFuncFromMvcMapping(m MvcMapping) gin.HandlerFunc {
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

	return MakeGinHandlerFunc(s, m.Condition())
}

func (r *Registrar) findMiddlewares(group, relativePath string, methods...string) (gin.HandlersChain, error) {
	var handlers = make([]gin.HandlerFunc, len(r.middlewares))
	sort.SliceStable(r.middlewares, func(i,j int) bool { return r.middlewares[i].Order() < r.middlewares[j].Order()})
	var i = 0
	path := NormalizedPath(relativePath)
	for _,mw := range r.middlewares {
		switch match, err := r.routeMatches(mw.Matcher(), group, path, methods...); {
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
	Helpers
***************************/
func MakeGinHandlerFunc(s *httptransport.Server, rm RequestMatcher) gin.HandlerFunc {
	handler := func(c *gin.Context) {
		reqCtx := utils.MakeMutableContext(c.Request.Context())
		reqCtx.Set(kGinContextKey, c)
		c.Request = c.Request.WithContext(reqCtx)
		s.ServeHTTP(c.Writer, c.Request)
	}
	return MakeConditionalHandlerFunc(handler, rm)
}

func MakeConditionalHandlerFunc(handler gin.HandlerFunc, rm RequestMatcher) gin.HandlerFunc {
	if rm == nil {
		return handler
	}
	return func(c *gin.Context) {
		if matches, e := rm.MatchesWithContext(c, c.Request); e == nil && matches {
			handler(c)
		}
	}
}

func ginContextExtractor(ctx context.Context, r *http.Request) (ret context.Context) {
	if ret = r.Context().Value(kGinContextKey).(context.Context); ret == nil {
		return ctx
	}
	return
}






