package matcher

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/matcher"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"fmt"
	pathutils "path"
)

// routeMatcher implement web.RouteMatcher
type routeMatcher struct {
	description string
	matchableFunc func(*web.Route) interface{}
	delegate matcher.Matcher
}

func (m *routeMatcher) RouteMatches(c context.Context, r *web.Route) (bool, error) {
	if m.matchableFunc == nil {
		return m.delegate.MatchesWithContext(c, r)
	}
	return m.delegate.MatchesWithContext(c, m.matchableFunc(r))
}

func (m *routeMatcher) Matches(i interface{}) (bool, error) {
	value, err := interfaceToRoute(i)
	if err != nil {
		return false, err
	}
	return m.RouteMatches(context.TODO(), value)
}

func (m *routeMatcher) MatchesWithContext(c context.Context, i interface{}) (bool, error) {
	value, err := interfaceToRoute(i)
	if err != nil {
		return false, err
	}
	return m.RouteMatches(c, value)
}

func (m *routeMatcher) Or(matchers ...matcher.Matcher) matcher.ChainableMatcher {
	return matcher.Or(m, matchers...)
}

func (m *routeMatcher) And(matchers ...matcher.Matcher) matcher.ChainableMatcher {
	return matcher.And(m, matchers...)
}

func (m *routeMatcher) String() string {
	switch stringer, ok :=m.delegate.(fmt.Stringer); {
	case len(m.description) != 0:
		return m.description
	case ok:
		return stringer.String()
	default:
		return "RouteMatcher"
	}
}

/**************************
	Constructors
***************************/
func AnyRoute() web.RouteMatcher {
	return wrapAsRouteMatcher(matcher.Any())
}

func RouteWithMethods(methods...string) web.RouteMatcher {
	var delegate matcher.ChainableMatcher
	if len(methods) == 0 {
		delegate = matcher.Any()
	} else {
		delegate = matcher.WithString(methods[0], true)
		for _,m := range methods[1:] {
			delegate = delegate.Or(matcher.WithString(m, true))
		}
	}

	return &routeMatcher{
		description: fmt.Sprintf("method %s", delegate.(fmt.Stringer).String()),
		matchableFunc: routeMethod,
		delegate: delegate,
	}
}

// PatternMatcher checks web.Route's path with prefix
// The prefix syntax is:
//
//  prefix:
//    { term }
//  term:
//    '*'         matches any sequence of non-path-separators
//    '**'        matches any sequence of characters, including
//                path separators.
//    '?'         matches any single non-path-separator character
//    '[' [ '^' ] { character-range } ']'
//          character class (must be non-empty)
//    '{' { term } [ ',' { term } ... ] '}'
//    c           matches character c (c != '*', '?', '\\', '[')
//    '\\' c      matches character c
//
//  character-range:
//    c           matches character c (c != '\\', '-', ']')
//    '\\' c      matches character c
//    lo '-' hi   matches character c for lo <= c <= hi
func RouteWithPattern(pattern string, methods...string) web.RouteMatcher {
	pDelegate := matcher.WithPathPattern(pattern)
	pMatcher := &routeMatcher{
		description: fmt.Sprintf("path %s", pDelegate.(fmt.Stringer).String()),
		matchableFunc: routeAbsPath,
		delegate: pDelegate,
	}
	mMatcher := RouteWithMethods(methods...)
	return wrapAsRouteMatcher(pMatcher.And(mMatcher))
}

func RouteWithPrefix(prefix string, methods...string) web.RouteMatcher {
	pDelegate := matcher.WithPrefix(prefix, true)
	pMatcher := &routeMatcher{
		description: fmt.Sprintf("path %s", pDelegate.(fmt.Stringer).String()),
		matchableFunc: routeAbsPath,
		delegate: pDelegate,
	}
	mMatcher := RouteWithMethods(methods...)
	return wrapAsRouteMatcher(pMatcher.And(mMatcher))
}

func RouteWithRegex(regex string, methods...string) web.RouteMatcher {
	pDelegate := matcher.WithRegex(regex)
	pMatcher := &routeMatcher{
		description: fmt.Sprintf("path %s", pDelegate.(fmt.Stringer).String()),
		matchableFunc: routeAbsPath,
		delegate: pDelegate,
	}
	mMatcher := RouteWithMethods(methods...)
	return wrapAsRouteMatcher(pMatcher.And(mMatcher))
}

func RouteWithGroup(group string) web.RouteMatcher {
	delegate := matcher.WithString(group, false)
	return &routeMatcher{
		description: fmt.Sprintf("group %s", delegate.(fmt.Stringer).String()),
		matchableFunc: routeGroup,
		delegate: delegate,
	}
}

/**************************
	helpers
***************************/
func interfaceToRoute(i interface{}) (*web.Route, error) {
	switch i.(type) {
	case web.Route:
		r := i.(web.Route)
		return &r, nil
	case *web.Route:
		return i.(*web.Route), nil
	default:
		return nil, fmt.Errorf("RouteMatcher doesn't support %T", i)
	}
}

func routeGroup(r *web.Route) interface{} {
	return r.Group
}

func routeMethod(r *web.Route) interface{} {
	return r.Method
}

func routeAbsPath(r *web.Route) interface{} {
	p := pathutils.Join(r.Group, r.Path)
	if !pathutils.IsAbs(p) {
		p = "/" + p
	}
	return pathutils.Clean(p)
}

func wrapAsRouteMatcher(m matcher.Matcher) web.RouteMatcher {
	return &routeMatcher{
		delegate: m,
	}
}