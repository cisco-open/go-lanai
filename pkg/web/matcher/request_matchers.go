package matcher

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/matcher"
	web "cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"fmt"
	"net/http"
	"strings"
)

// requestMatcher implement web.RequestMatcher
type requestMatcher struct {
	description   string
	matchableFunc func(context.Context, *http.Request) (interface{}, error)
	delegate      matcher.Matcher
}

func (m *requestMatcher) RequestMatches(c context.Context, r *http.Request) (bool, error) {
	if m.matchableFunc == nil {
		return m.delegate.MatchesWithContext(c, r)
	}
	matchable, err := m.matchableFunc(c, r)
	if err != nil {
		return false, err
	}
	return m.delegate.MatchesWithContext(c, matchable)
}

func (m *requestMatcher) Matches(i interface{}) (bool, error) {
	value, err := interfaceToRequest(i)
	if err != nil {
		return false, err
	}
	return m.RequestMatches(context.TODO(), value)
}

func (m *requestMatcher) MatchesWithContext(c context.Context, i interface{}) (bool, error) {
	value, err := interfaceToRequest(i)
	if err != nil {
		return false, err
	}
	return m.RequestMatches(c, value)
}

func (m *requestMatcher) Or(matchers ...matcher.Matcher) matcher.ChainableMatcher {
	return matcher.Or(m, matchers...)
}

func (m *requestMatcher) And(matchers ...matcher.Matcher) matcher.ChainableMatcher {
	return matcher.And(m, matchers...)
}

func (m *requestMatcher) String() string {
	switch stringer, ok :=m.delegate.(fmt.Stringer); {
	case len(m.description) != 0:
		return m.description
	case ok:
		return stringer.String()
	default:
		return "web.RequestMatcher"
	}
}

/**************************
	Constructors
***************************/
func AnyRequest() web.RequestMatcher {
	return wrapAsRequestMatcher(matcher.Any())
}

// TODO support wildcard
func RequestWithHost(expected string) web.RequestMatcher {
	delegate := matcher.WithString(expected, true)
	return &requestMatcher{
		description:   fmt.Sprintf("host %s", delegate.(fmt.Stringer).String()),
		matchableFunc: host,
		delegate:      delegate,
	}
}

func RequestWithMethods(methods...string) web.RequestMatcher {

	var delegate matcher.ChainableMatcher
	if len(methods) == 0 {
		delegate = matcher.Any()
	} else {
		delegate = matcher.WithString(methods[0], true)
		for _,m := range methods[1:] {
			delegate = delegate.Or(matcher.WithString(m, true))
		}
	}

	return &requestMatcher{
		description:   fmt.Sprintf("method %s", delegate.(fmt.Stringer).String()),
		matchableFunc: method,
		delegate:      delegate,
	}
}

// RequestWithPattern create a web.RequestMatcher with path pattern.
// if context is available when performing the match, the context path is striped
func RequestWithPattern(pattern string, methods...string) web.RequestMatcher {
	pDelegate := matcher.WithPathPattern(pattern)
	pMatcher := &requestMatcher{
		description:   fmt.Sprintf("path %s", pDelegate.(fmt.Stringer).String()),
		matchableFunc: path,
		delegate:      pDelegate,
	}
	mMatcher := RequestWithMethods(methods...)
	return wrapAsRequestMatcher(pMatcher.And(mMatcher))
}

// RequestWithPrefix create a web.RequestMatcher with prefix
// if context is available when performing the match, the context path is striped
func RequestWithPrefix(prefix string, methods...string) web.RequestMatcher {
	pDelegate := matcher.WithPrefix(prefix, true)
	pMatcher := &requestMatcher{
		description:   fmt.Sprintf("path %s", pDelegate.(fmt.Stringer).String()),
		matchableFunc: path,
		delegate:      pDelegate,
	}
	mMatcher := RequestWithMethods(methods...)
	return wrapAsRequestMatcher(pMatcher.And(mMatcher))
}

// RequestWithPrefix create a web.RequestMatcher with regular expression
// if context is available when performing the match, the context path is striped
func RequestWithRegex(regex string, methods...string) web.RequestMatcher {
	pDelegate := matcher.WithRegex(regex)
	pMatcher := &requestMatcher{
		description:   fmt.Sprintf("path %s", pDelegate.(fmt.Stringer).String()),
		matchableFunc: path,
		delegate:      pDelegate,
	}
	mMatcher := RequestWithMethods(methods...)

	return wrapAsRequestMatcher(pMatcher.And(mMatcher))
}

// TODO more request matchers

/**************************
	helpers
***************************/
func interfaceToRequest(i interface{}) (*http.Request, error) {
	switch i.(type) {
	case http.Request:
		r := i.(http.Request)
		return &r, nil
	case *http.Request:
		return i.(*http.Request), nil
	default:
		return nil, fmt.Errorf("web.RequestMatcher doesn't support %T", i)
	}
}

func wrapAsRequestMatcher(m matcher.Matcher) web.RequestMatcher {
	var desc string
	if stringer, ok := m.(fmt.Stringer); ok {
		desc = stringer.String()
	}
	return &requestMatcher{
		description: desc,
		delegate: m,
	}
}

func host(_ context.Context, r *http.Request) (interface{}, error) {
	return r.Host, nil
}

func method(_ context.Context, r *http.Request) (interface{}, error) {
	return r.Method, nil
}

func path(c context.Context, r *http.Request) (interface{}, error) {
	path := r.URL.Path
	ctxPath, ok := c.Value(web.ContextKeyContextPath).(string)
	if !ok {
		return nil, fmt.Errorf("context is required to match request path")
	}
	return strings.TrimPrefix(path, ctxPath), nil
}