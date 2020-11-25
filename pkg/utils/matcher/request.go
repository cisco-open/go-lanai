package matcher

import (
	"fmt"
	"net/http"
)

// RequestMatcher is a typed ChainableMatcher that accept *http.Request or http.Request
type RequestMatcher interface {
	ChainableMatcher
	RequestMatches(*http.Request) (bool, error)
}

// requestMatcher implement RequestMatcher
type requestMatcher struct {
	description string
	matchableFunc func(r *http.Request) interface{}
	delegate Matcher
}

func (m *requestMatcher) RequestMatches(r *http.Request) (bool, error) {
	if m.matchableFunc == nil {
		return m.delegate.Matches(r)
	}
	return m.delegate.Matches(m.matchableFunc(r))
}

func (m *requestMatcher) Matches(i interface{}) (bool, error) {
	var value *http.Request
	switch i.(type) {
	case http.Request:
		r := i.(http.Request)
		value = &r
	case *http.Request:
		value = i.(*http.Request)
	default:
		return false, fmt.Errorf("RequestMatcher doesn't support %T", i)
	}
	return m.RequestMatches(value)

}

func (m *requestMatcher) Or(matchers ...Matcher) ChainableMatcher {
	return Or(m, matchers...)
}

func (m *requestMatcher) And(matchers ...Matcher) ChainableMatcher {
	return And(m, matchers...)
}

func (m *requestMatcher) String() string {
	switch stringer, ok :=m.delegate.(fmt.Stringer); {
	case len(m.description) != 0:
		return m.description
	case ok:
		return stringer.String()
	default:
		return "RequestMatcher"
	}
}

/**************************
	Constructors
***************************/
func AnyRequest() RequestMatcher {
	return wrapAsRequestMatcher(Any())
}

// TODO support wildcard
func RequestWithHost(expected string) RequestMatcher {
	delegate := WithString(expected, true)
	return &requestMatcher{
		description: fmt.Sprintf("host %s", delegate.(fmt.Stringer).String()),
		matchableFunc: host,
		delegate: delegate,
	}
}

func RequestWithMethods(methods...string) RequestMatcher {

	var delegate ChainableMatcher
	if len(methods) == 0 {
		delegate = Any()
	} else {
		delegate = WithString(methods[0], true)
		for _,m := range methods[1:] {
			delegate = delegate.Or(WithString(m, true))
		}
	}

	return &requestMatcher{
		description: fmt.Sprintf("method %s", delegate.(fmt.Stringer).String()),
		matchableFunc: method,
		delegate: delegate,
	}
}

func RequestWithPattern(pattern string, methods...string) RequestMatcher {
	pDelegate := WithPathPattern(pattern)
	pMatcher := &requestMatcher{
		description: fmt.Sprintf("path %s", pDelegate.(fmt.Stringer).String()),
		matchableFunc: path,
		delegate: pDelegate,
	}
	mMatcher := RequestWithMethods(methods...)
	return wrapAsRequestMatcher(pMatcher.And(mMatcher))
}

func RequestWithPrefix(prefix string, methods...string) RequestMatcher {
	pDelegate := WithPrefix(prefix, true)
	pMatcher := &requestMatcher{
		description: fmt.Sprintf("path %s", pDelegate.(fmt.Stringer).String()),
		matchableFunc: path,
		delegate: pDelegate,
	}
	mMatcher := RequestWithMethods(methods...)
	return wrapAsRequestMatcher(pMatcher.And(mMatcher))
}

func RequestWithRegex(regex string, methods...string) RequestMatcher {
	pDelegate := WithRegex(regex)
	pMatcher := &requestMatcher{
		description: fmt.Sprintf("path %s", pDelegate.(fmt.Stringer).String()),
		matchableFunc: path,
		delegate: pDelegate,
	}
	mMatcher := RequestWithMethods(methods...)

	return wrapAsRequestMatcher(pMatcher.And(mMatcher))
}

// TODO more request matchers

/**************************
	helpers
***************************/
func wrapAsRequestMatcher(m Matcher) RequestMatcher {
	return &requestMatcher{
		delegate: m,
	}
}

func host(r *http.Request) interface{} {
	return r.Host
}

func method(r *http.Request) interface{} {
	return r.Method
}

func path(r *http.Request) interface{} {
	//TODO stripe context path
	return r.URL.Path
}