package web

type ConditionalMiddleware interface {
	Condition() RequestMatcher
}

type Middleware interface {
	HandlerFunc() HandlerFunc
}

type middlewareMapping struct {
	name        string
	order       int
	matcher     RouteMatcher
	condition   RequestMatcher
	handlerFunc HandlerFunc
}

func NewMiddlewareMapping(name string, order int, matcher RouteMatcher, cond RequestMatcher, handlerFunc HandlerFunc) MiddlewareMapping {
	return &middlewareMapping {
		name: name,
		matcher: matcher,
		order: order,
		condition: cond,
		handlerFunc: handlerFunc,
	}
}

func (mm middlewareMapping) Name() string {
	return mm.name
}

func (mm middlewareMapping) Matcher() RouteMatcher {
	return mm.matcher
}

func (mm middlewareMapping) Order() int {
	return mm.order
}

func (mm middlewareMapping) Condition() RequestMatcher {
	return mm.condition
}

func (mm middlewareMapping) HandlerFunc() HandlerFunc {
	return mm.handlerFunc
}



