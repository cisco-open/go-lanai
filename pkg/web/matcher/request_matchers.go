package matcher

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/matcher"
	web "cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

const (
	descTmplPath = `path %s`
)

type matchableFunc func(context.Context, *http.Request) (interface{}, error)

// requestMatcher implement web.RequestMatcher
type requestMatcher struct {
	description   string
	matchableFunc matchableFunc
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

func NoneRequest() web.RequestMatcher {
	return wrapAsRequestMatcher(matcher.None())
}

func NotRequest(m web.RequestMatcher) web.RequestMatcher {
	return wrapAsRequestMatcher(matcher.Not(m))
}

// RequestWithHost
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
		description:   fmt.Sprintf("method %v", delegate),
		matchableFunc: method,
		delegate:      delegate,
	}
}

// RequestWithPattern create a web.RequestMatcher with path pattern.
// if context is available when performing the match, the context path is striped
func RequestWithPattern(pattern string, methods...string) web.RequestMatcher {
	pDelegate := matcher.WithPathPattern(pattern)
	pMatcher := &requestMatcher{
		description:   fmt.Sprintf(descTmplPath, pDelegate.(fmt.Stringer).String()),
		matchableFunc: path,
		delegate:      pDelegate,
	}
	mMatcher := RequestWithMethods(methods...)
	return wrapAsRequestMatcher(pMatcher.And(mMatcher))
}

// RequestWithURL is similar with RequestWithPattern, but instead it takes a relative URL path and convert it to pattern
// by extracting "path" part (remove #fragment, ?query and more)
func RequestWithURL(url string, methods...string) web.RouteMatcher {
	return RequestWithPattern(PatternFromURL(url), methods...)
}

// RequestWithPrefix create a web.RequestMatcher with prefix
// if context is available when performing the match, the context path is striped
func RequestWithPrefix(prefix string, methods...string) web.RequestMatcher {
	pDelegate := matcher.WithPrefix(prefix, true)
	pMatcher := &requestMatcher{
		description:   fmt.Sprintf(descTmplPath, pDelegate.(fmt.Stringer).String()),
		matchableFunc: path,
		delegate:      pDelegate,
	}
	mMatcher := RequestWithMethods(methods...)
	return wrapAsRequestMatcher(pMatcher.And(mMatcher))
}

// RequestWithRegex create a web.RequestMatcher with regular expression
// if context is available when performing the match, the context path is striped
func RequestWithRegex(regex string, methods...string) web.RequestMatcher {
	pDelegate := matcher.WithRegex(regex)
	pMatcher := &requestMatcher{
		description:   fmt.Sprintf(descTmplPath, pDelegate.(fmt.Stringer).String()),
		matchableFunc: path,
		delegate:      pDelegate,
	}
	mMatcher := RequestWithMethods(methods...)

	return wrapAsRequestMatcher(pMatcher.And(mMatcher))
}

func RequestWithHeader(name string, value string, prefix bool) web.RequestMatcher {
	matchable := func(_ context.Context, r *http.Request) (interface{}, error) {
		return r.Header.Get(name), nil
	}

	var delegate matcher.Matcher

	if prefix {
		delegate = matcher.WithPrefix(value, true)
	} else {
		delegate = matcher.WithString(value, true)
	}

	return &requestMatcher{
		description: fmt.Sprintf("matches header %s:%s", name, value),
		matchableFunc: matchable,
		delegate: delegate,
	}
}

func RequestHasHeader(name string) web.RequestMatcher {
	matchable := func(_ context.Context, r *http.Request) (interface{}, error) {
		return r.Header.Get(name), nil
	}
	return &requestMatcher{
		description: fmt.Sprintf("matches have header %s", name),
		matchableFunc: matchable,
		delegate: matcher.AnyNonEmptyString(),
	}
}

func RequestHasPostParameter(name string) web.RequestMatcher {
	matchable := func(ctx context.Context, r *http.Request) (interface{}, error) {
		if g := web.GinContext(ctx); g != nil {
			p, _ := g.GetPostForm(name)
			return p, nil
		} else if _ = r.ParseForm(); len(r.PostForm[name]) != 0 {
			return r.PostForm[name][0], nil
		}
		return nil, errors.New("can't get post param from context. *gin.Context is not available")
	}

	return &requestMatcher{
		description: fmt.Sprintf("matches have parameter %s", name),
		matchableFunc: matchable,
		delegate: matcher.AnyNonEmptyString(),
	}
}

func RequestWithParam(name, value string) web.RequestMatcher {
	return &requestMatcher{
		description: fmt.Sprintf("matches have parameter %s", name),
		matchableFunc: param(name),
		delegate: matcher.WithString(value, true),
	}
}

func CustomMatcher(description string, matchable matchableFunc, delegate matcher.Matcher ) web.RequestMatcher {
	return &requestMatcher{
		description: description,
		matchableFunc: matchable,
		delegate: delegate,
	}
}

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

func param(name string) matchableFunc {
	return func (c context.Context, r *http.Request) (interface{}, error) {
		if e := r.ParseForm(); e != nil {
			return nil, e
		}
		return r.Form.Get(name), nil
	}
}

