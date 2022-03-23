package web

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/order"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/reflectutils"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	httptransport "github.com/go-kit/kit/transport/http"
	"go.uber.org/fx"
	"html/template"
	"io/fs"
	"net"
	"net/http"
	pathutils "path"
	"reflect"
	"sort"
	"strings"
	"time"
)

//goland:noinspection GoUnusedConst
const (
	kGinContextKey = "GinCtx"
	kKitContextKey = "KitCtx"
	DefaultGroup   = "/"
)

type Registrar struct {
	engine           *Engine
	router           gin.IRouter
	server           *http.Server
	port             int
	properties       ServerProperties
	options          []*orderedServerOption // options go-kit server options
	validator        *Validate
	requestRewriter  RequestRewriter
	middlewares      []MiddlewareMapping // middlewares gin-gonic middleware providers
	routedMappings   routedMappings      // routedMappings MvcMappings + SimpleMappings
	staticMappings   []StaticMapping     // staticMappings all static mappings
	customizers      []Customizer
	errMappings      []ErrorTranslateMapping
	errTranslators   []ErrorTranslator
	embedFs          []fs.FS
	initialized      bool
	warnDuplicateMWs bool
	warnExclusion    utils.StringSet
}

func NewRegistrar(g *Engine, properties ServerProperties) *Registrar {

	var contextPath = pathutils.Clean("/" + properties.ContextPath)
	registrar := &Registrar{
		engine:     g,
		router:     g.Group(contextPath),
		properties: properties,
		options: []*orderedServerOption{
			newOrderedServerOption(httptransport.ServerBefore(integrateGinContextBefore), order.Lowest),
			newOrderedServerOption(httptransport.ServerFinalizer(integrateGinContextFinalizer), order.Lowest),
		},
		validator:        bindingValidator,
		requestRewriter:  newGinRequestRewriter(g.Engine),
		routedMappings:   routedMappings{},
		warnDuplicateMWs: true,
		warnExclusion:    utils.NewStringSet(),
	}
	return registrar
}

// initialize should be called during application startup, last change to change configurations, load templates, etc
func (r *Registrar) initialize(ctx context.Context) (err error) {
	if r.initialized {
		return fmt.Errorf("attempting to initialize web engine multiple times")
	}

	// first, we add some manditory customizers and middleware
	r.MustRegister(NewPriorityGinContextCustomizer(&r.properties))
	r.MustRegister(NewGinContextCustomizer(&r.properties))

	// apply customizers before install mappings
	if err = r.applyCustomizers(ctx); err != nil {
		return
	}

	// we disable auto-validation. We will invoke our own validation manually.
	// Also we need to make the validator available globally for any request decoder to access.
	// The alternative approach is to put the validator into each gin.Context
	binding.Validator = nil

	// load templates
	r.loadHtmlTemplates(ctx)

	// add some common middlewares
	var mappings []interface{}
	if err = r.Register(mappings...); err != nil {
		return
	}

	// before starting to register mappings, we want global MW to take effect on our main group
	var contextPath = pathutils.Clean("/" + r.properties.ContextPath)
	r.router = r.engine.Group(contextPath)

	// register routedMappings to gin engine
	if err = r.installMappings(ctx); err != nil {
		return
	}

	r.initialized = true
	return
}

// cleanup post initilaize cleanups
func (r *Registrar) cleanup(ctx context.Context) (err error) {
	if e := r.applyPostInitCustomizers(ctx); e != nil {
		return e
	}
	return nil
}

// AddGlobalMiddlewares add middleware to all mapping
func (r *Registrar) AddGlobalMiddlewares(handlerFuncs ...gin.HandlerFunc) error {
	r.engine.Use(handlerFuncs...)
	return nil
}

// AddOption calls AddOptionWithOrder with 0 order value
func (r *Registrar) AddOption(opt httptransport.ServerOption) error {
	return r.AddOptionWithOrder(opt, 0)
}

// AddOptionWithOrder add go-kit ServerOption with order.
// httptransport.ServerOption are ordered using order.OrderedFirstCompare
func (r *Registrar) AddOptionWithOrder(opt httptransport.ServerOption, o int) error {
	if r.initialized {
		return fmt.Errorf("cannot register options after web engine have initialized")
	}

	r.options = append(r.options, newOrderedServerOption(opt, o))
	order.SortStable(r.options, order.OrderedFirstCompare)
	return nil
}

// AddEngineOptions customize Engine
func (r *Registrar) AddEngineOptions(opts ...EngineOptions) error {
	for _, fn := range opts {
		fn(r.engine)
	}
	return nil
}

func (r *Registrar) WarnDuplicateMiddlewares(ifWarn bool, excludedPath ...string) {
	r.warnDuplicateMWs = ifWarn
	r.warnExclusion.Add(excludedPath...)
}

// Run configure and start gin engine
func (r *Registrar) Run(ctx context.Context) (err error) {
	if err = r.initialize(ctx); err != nil {
		return
	}
	defer func(ctx context.Context) {
		_ = r.cleanup(ctx)
	}(ctx)

	// we let system to choose port if not set
	var addr = fmt.Sprintf(":%v", r.properties.Port)
	if r.properties.Port <= 0 {
		addr = ":0"
	}

	r.server = &http.Server{
		Addr:           addr,
		Handler:        r.engine,
		ReadTimeout:    60 * time.Second,
		WriteTimeout:   60 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	// start the server
	tcpAddr, e := r.listenAndServe()
	if e == nil {
		r.port = tcpAddr.Port
	}
	return e
}

// Stop closes http server
func (r *Registrar) Stop(ctx context.Context) (err error) {
	if r.server == nil {
		return fmt.Errorf("attempt to stop server before initialization")
	}
	err = r.server.Close()
	if err != nil {
		logger.WithContext(ctx).Warnf("error when stop http server: %v", err)
	} else {
		logger.WithContext(ctx).Infof("http server stopped")
	}
	return
}

// ServerPort returns the port of started server, returns 0 if server is not initialized
func (r *Registrar) ServerPort() int {
	return r.port
}

// Register is the entry point to register Controller, Mapping and other web related objects
// supported items type are:
// 	- Customizer
// 	- Controller
//  - EndpointMapping
//  - StaticMapping
//  - TemplateMapping
//  - MiddlewareMapping
//  - ErrorTranslateMapping
//  - ErrorTranslator
//  - struct that contains exported Controller fields
//  - fs.FS
func (r *Registrar) Register(items ...interface{}) (err error) {
	for _, i := range items {
		if err = r.register(i); err != nil {
			break
		}
	}
	return
}

func (r *Registrar) MustRegister(items ...interface{}) {
	if e := r.Register(items...); e != nil {
		panic(e)
	}
}

// RegisterWithLifecycle is a convenient function to schedule item registration in FX lifecycle
func (r *Registrar) RegisterWithLifecycle(lc fx.Lifecycle, items ...interface{}) {
	lc.Append(fx.Hook{
		OnStart: func(context.Context) (err error) {
			return r.Register(items...)
		},
	})
}

func (r *Registrar) listenAndServe() (*net.TCPAddr, error) {
	ln, err := net.Listen("tcp", r.server.Addr)
	if err != nil {
		return nil, err
	}

	go func() {
		_ = r.server.Serve(ln)
	}()
	return ln.Addr().(*net.TCPAddr), nil
}

func (r *Registrar) register(i interface{}) (err error) {
	if r.initialized {
		return errors.New("attempting to register mappings/middlewares/pre-processors after web engine initialization")
	}

	switch v := i.(type) {
	case Controller:
		err = r.registerController(v)
	case MvcMapping:
		err = r.registerMvcMapping(v)
	case StaticMapping:
		err = r.registerStaticMapping(v)
	case MiddlewareMapping:
		err = r.registerMiddlewareMapping(v)
	case SimpleMapping:
		err = r.registerSimpleMapping(v)
	case ErrorTranslateMapping:
		err = r.registerErrorMapping(v)
	case RequestPreProcessor:
		err = r.registerRequestPreProcessor(v)
	case Customizer:
		err = r.registerWebCustomizer(v)
	case ErrorTranslator:
		err = r.registerErrorTranslator(v)
	case fs.FS:
		r.embedFs = append(r.embedFs, v)
	default:
		err = r.registerUnknownType(i)
	}
	return
}

func (r *Registrar) registerUnknownType(i interface{}) (err error) {
	v := reflect.ValueOf(i)
	for ; v.Kind() == reflect.Ptr; v = v.Elem() {
		// SuppressWarnings go:S108 empty block is intended
	}

	var valid bool
	switch {
	case v.Kind() == reflect.Slice:
		for i := 0; i < v.Len(); i++ {
			if e := r.register(v.Index(i).Interface()); e != nil {
				return e
			}
		}
		// empty slice doesn't count as error
		valid = true
	case v.Kind() == reflect.Struct:
		// go through fields and register
		for idx := 0; idx < v.NumField(); idx++ {
			// only care controller fields
			if f := v.Type().Field(idx); !reflectutils.IsExportedField(f) {
				// unexported field
				continue
			}
			c := v.Field(idx).Interface()
			switch c.(type) {
			case fx.In:
				valid = true
			case Controller:
				valid = true
				if e := r.register(c); e != nil {
					return e
				}
			}
		}
	}
	if !valid {
		return errors.New(fmt.Sprintf("unsupported type [%T]", i))
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
	group := DefaultGroup
	if m.Group() != "" {
		group = m.Group()
	}

	paths := r.routedMappings.GetOrNew(method).GetOrNew(group)
	mappings := paths.GetOrNew(path)
	mappings = append(mappings, m)
	paths[path] = mappings
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
	order.SortStable(r.customizers, order.OrderedFirstCompare)
	return nil
}

func (r *Registrar) registerErrorMapping(m ErrorTranslateMapping) error {
	if r.initialized {
		return fmt.Errorf("cannot register error mappings after web engine have initialized")
	}
	r.errMappings = append(r.errMappings, m)
	return nil
}

func (r *Registrar) registerErrorTranslator(t ErrorTranslator) error {
	if r.initialized {
		return fmt.Errorf("cannot register error translator after web engine have initialized")
	}
	r.errTranslators = append(r.errTranslators, t)
	return nil
}

func (r *Registrar) applyCustomizers(ctx context.Context) error {
	if r.customizers == nil {
		return nil
	}
	for _, c := range r.customizers {
		if e := c.Customize(ctx, r); e != nil {
			return e
		}
	}
	return nil
}

func (r *Registrar) applyPostInitCustomizers(ctx context.Context) error {
	if r.customizers == nil {
		return nil
	}
	for _, c := range r.customizers {
		if pi, ok := c.(PostInitCustomizer); ok {
			if e := pi.PostInit(ctx, r); e != nil {
				return e
			}
		}
	}
	return nil
}

func (r *Registrar) installMappings(ctx context.Context) error {
	// before registering, we need to add default error translators
	order.SortStable(r.errMappings, order.OrderedFirstCompare)
	r.errTranslators = append(r.errTranslators, newDefaultErrorTranslator())

	// register routedMappings
	for method, groups := range r.routedMappings {
		for group, paths := range groups {
			for _, mappings := range paths {
				// all routedMappings with condition registered first
				sort.SliceStable(mappings, func(i, j int) bool {
					return mappings[i].Condition() != nil && mappings[j].Condition() == nil
				})

				if e := r.installRoutedMappings(ctx, method, group, mappings); e != nil {
					return e
				}
			}
		}

	}

	// register static mappings
	for _, m := range r.staticMappings {
		if e := r.installStaticMapping(ctx, m); e != nil {
			return e
		}
	}
	return nil
}

func (r *Registrar) installStaticMapping(ctx context.Context, m StaticMapping) error {
	embedded := make([]fs.FS, len(r.embedFs))
	for i, fsys := range r.embedFs {
		embedded[i] = OrderedFS(NewDirFS(m.StaticRoot(), fsys), i)
	}

	mFs := NewMergedFS(OrderedFS(NewOSDirFS(m.StaticRoot()), order.Highest), embedded...)
	mw := ginStaticAssetsHandler{
		rewriter: r.requestRewriter,
		fsys:     mFs,
		aliases:  m.Aliases(),
	}

	middlewares, err := r.findMiddlewares(ctx, DefaultGroup, m.Path(), http.MethodGet, http.MethodHead)
	middlewares = append(gin.HandlersChain{mw.FilenameRewriteHandlerFunc()}, middlewares...)
	middlewares = append(middlewares, mw.PreCompressedGzipAsset())
	r.router.Group(DefaultGroup).
		Use(middlewares...).
		StaticFS(m.Path(), http.FS(mFs))
	return err
}

//nolint:contextcheck // context is only for logging purpose
func (r *Registrar) installRoutedMappings(ctx context.Context, method, group string, mappings []RoutedMapping) error {
	if len(mappings) == 0 {
		return nil
	}
	if group == "" {
		group = DefaultGroup
	}

	path := mappings[0].Path()
	// resolve error translators first
	errTranslators, e := r.findErrorTranslators(ctx, group, path, method)
	if e != nil {
		return fmt.Errorf("unable to resolve error translation for [%s %s]", method, path)
	}
	order.SortStable(errTranslators, order.OrderedFirstCompare)

	// resolve gin.HandlerFunc to register
	handlerFuncs := make([]gin.HandlerFunc, len(mappings))
	options := r.kitServerOptions()
	unconditionalFound := false
	for i, m := range mappings {
		// validate method and path with best efforts
		switch {
		case path != m.Path():
			return fmt.Errorf("attempt to register multiple RoutedMappings with inconsist path parameters: "+
				"expected [%s (%s)%s] but got [%s (%s)%s]", method, group, path, m.Method(), m.Group(), m.Path())
		case m.Condition() == nil && unconditionalFound:
			return fmt.Errorf("attempt to register multiple unconditional RoutedMappings on same path and method: [%s %s]", m.Method(), m.Path())
		case m.Condition() == nil:
			unconditionalFound = true
		}

		// create handler func
		switch m.(type) {
		case MvcMapping:
			handlerFuncs[i] = r.makeHandlerFuncFromMvcMapping(m.(MvcMapping), errTranslators, options)
		case SimpleGinMapping:
			handlerFuncs[i] = r.makeGinConditionalHandlerFunc(m.(SimpleGinMapping).GinHandlerFunc(), m.Condition())
		case SimpleMapping:
			f := NewHttpGinHandlerFunc(http.HandlerFunc(m.(SimpleMapping).HandlerFunc()))
			handlerFuncs[i] = r.makeGinConditionalHandlerFunc(f, m.Condition())
		}
	}

	// find middleware and register with router
	middlewares, err := r.findMiddlewares(ctx, group, path, method)
	if method == MethodAny {
		r.router.Group(group).
			Use(middlewares...).
			Any(path, handlerFuncs...)
	} else {
		r.router.Group(group).
			Use(middlewares...).
			Handle(method, path, handlerFuncs...)
	}

	return err
}

//nolint:contextcheck // context is only for logging purpose
func (r *Registrar) findMiddlewares(ctx context.Context, group, relativePath string, methods ...string) (gin.HandlersChain, error) {
	var handlers = make([]gin.HandlerFunc, len(r.middlewares))
	var matchedMW = make([]MiddlewareMapping, len(r.middlewares))
	sort.SliceStable(r.middlewares, func(i, j int) bool { return r.middlewares[i].Order() < r.middlewares[j].Order() })
	var i = 0
	path := NormalizedPath(relativePath)
	for _, mw := range r.middlewares {
		switch match, err := r.routeMatches(mw.Matcher(), group, path, methods...); {
		case err != nil:
			return []gin.HandlerFunc{}, err
		case match:
			var f gin.HandlerFunc
			switch mw.(type) {
			case MiddlewareGinMapping:
				f = mw.(MiddlewareGinMapping).GinHandlerFunc()
			default:
				f = NewHttpGinHandlerFunc(http.HandlerFunc(mw.HandlerFunc()))
			}
			handlers[i] = r.makeGinConditionalHandlerFunc(f, mw.Condition())
			matchedMW[i] = mw
			i++
		}
	}
	// warn duplicate MWs
	if r.warnDuplicateMWs {
		r.logMatchedMiddlewares(ctx, matchedMW[:i], group, relativePath, methods)
	}
	return handlers[:i], nil
}

func (r *Registrar) findErrorTranslators(_ context.Context, group, relativePath string, methods ...string) ([]ErrorTranslator, error) {
	var translators = make([]ErrorTranslator, len(r.errTranslators), len(r.errTranslators)+len(r.errMappings))
	for i, t := range r.errTranslators {
		translators[i] = t
	}

	var matched = make([]ErrorTranslateMapping, len(r.errMappings))
	path := NormalizedPath(relativePath)
	for i, m := range r.errMappings {
		switch match, err := r.routeMatches(m.Matcher(), group, path, methods...); {
		case err != nil:
			return translators, err
		case match:
			matched[i] = m
			translators = append(translators, newMappedErrorTranslator(m))
		}
	}
	return translators, nil
}

func (r *Registrar) routeMatches(matcher RouteMatcher, group, relativePath string, methods ...string) (bool, error) {
	switch {
	case len(methods) == 0:
		return false, fmt.Errorf("unable to register middleware: method is missing for %s", relativePath)
	case matcher == nil:
		return true, nil // no matcher, any value is a match
	}

	// match if any given method matches
	for _, m := range methods {
		ret, err := matcher.Matches(Route{Group: group, Path: relativePath, Method: m})
		if ret || err != nil {
			return ret, err
		}
	}
	return false, nil
}

func (r *Registrar) kitServerOptions() []httptransport.ServerOption {
	opts := make([]httptransport.ServerOption, len(r.options))
	for i, opt := range r.options {
		opts[i] = opt.ServerOption
	}
	return opts
}

func (r *Registrar) loadHtmlTemplates(ctx context.Context) {
	osFS := NewOSDirFS("web/", DirFSAllowListDirectory)
	mFs := NewMergedFS(osFS, r.embedFs...)
	t, e := template.ParseFS(mFs, "**/*.tmpl")
	if e != nil {
		logger.WithContext(ctx).Infof("no templates loaded: %v", e)
		return
	}
	r.engine.SetHTMLTemplate(t)
}

/**************************
	Helpers
***************************/

func (r *Registrar) makeHandlerFuncFromMvcMapping(m MvcMapping, errTranslators []ErrorTranslator, options []httptransport.ServerOption) gin.HandlerFunc {
	// create error encoder
	errenc := m.ErrorEncoder()
	if errenc == nil {
		errenc = JsonErrorEncoder()
	}

	options = append(options, httptransport.ServerErrorEncoder(
		newErrorEncoder(errenc, errTranslators...),
	))

	s := httptransport.NewServer(
		m.Endpoint(),
		m.DecodeRequestFunc(),
		m.EncodeResponseFunc(),
		options...,
	)

	return r.makeGinConditionalHandlerFunc(NewKitGinHandlerFunc(s), m.Condition()) //nolint:contextcheck
}

// makeGinConditionalHandlerFunc wraps given handler with a request matcher
func (r *Registrar) makeGinConditionalHandlerFunc(handler gin.HandlerFunc, rm RequestMatcher) gin.HandlerFunc {
	if rm == nil {
		return handler
	}
	return func(c *gin.Context) {
		if matches, e := rm.MatchesWithContext(c, c.Request); e == nil && matches {
			handler(c)
		} else if e != nil {
			_ = c.Error(e)
			c.Abort()
		}
	}
}

// logMatchedMiddlewares logs important information about middlewares, majorly for debug and early error detecting purpose
func (r *Registrar) logMatchedMiddlewares(ctx context.Context, matched []MiddlewareMapping, group, path string, methods []string) {
	// for now, we only warn about duplicates
	seen := map[string][]MiddlewareMapping{}
	for _, mw := range matched {
		if mw.Name() == "" {
			continue
		}
		v := seen[mw.Name()]
		v = append(v, mw)
		seen[mw.Name()] = v
	}

	var dups []string
	for k, v := range seen {
		if len(v) <= 1 || r.warnExclusion.Has(path) {
			continue
		}
		dups = append(dups, fmt.Sprintf(`"%s"x%d`, k, len(v)))
	}
	if len(dups) > 0 {
		if group == "/" {
			group = ""
		}
		logger.WithContext(ctx).Warnf("multiple Middlewares with same name detected at %s%s %v: %v", group, path, methods, dups)
	}

	return
}

/***************************
	helper types
 ***************************/

type routedMappings map[string]groupsMap

func (m routedMappings) GetOrNew(key string) groupsMap {
	if v, ok := m[key]; ok {
		return v
	}

	v := groupsMap{}
	m[key] = v
	return v
}

type groupsMap map[string]pathsMap

func (m groupsMap) GetOrNew(key string) pathsMap {
	if v, ok := m[key]; ok {
		return v
	}

	v := pathsMap{}
	m[key] = v
	return v
}

type pathsMap map[string][]RoutedMapping

func (m pathsMap) GetOrNew(key string) []RoutedMapping {
	if v, ok := m[key]; ok {
		return v
	}

	v := make([]RoutedMapping, 0)
	m[key] = v
	return v
}
