// Copyright 2023 Cisco Systems, Inc. and its affiliates
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package matcher

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/matcher"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"fmt"
	"net/url"
	pathutils "path"
	"strings"
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

// RouteWithPattern checks web.Route's path with prefix
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
		description: fmt.Sprintf(descTmplPath, pDelegate.(fmt.Stringer).String()),
		matchableFunc: routeAbsPath,
		delegate: pDelegate,
	}
	mMatcher := RouteWithMethods(methods...)
	return wrapAsRouteMatcher(pMatcher.And(mMatcher))
}

// RouteWithURL is similar with RouteWithPattern, but instead it takes a relative URL path and convert it to pattern
// by extracting "path" part (remove #fragment, ?query and more)
func RouteWithURL(url string, methods...string) web.RouteMatcher {
	return RouteWithPattern(PatternFromURL(url), methods...)
}

func RouteWithPrefix(prefix string, methods...string) web.RouteMatcher {
	pDelegate := matcher.WithPrefix(prefix, true)
	pMatcher := &routeMatcher{
		description: fmt.Sprintf(descTmplPath, pDelegate.(fmt.Stringer).String()),
		matchableFunc: routeAbsPath,
		delegate: pDelegate,
	}
	mMatcher := RouteWithMethods(methods...)
	return wrapAsRouteMatcher(pMatcher.And(mMatcher))
}

func RouteWithRegex(regex string, methods...string) web.RouteMatcher {
	pDelegate := matcher.WithRegex(regex)
	pMatcher := &routeMatcher{
		description: fmt.Sprintf(descTmplPath, pDelegate.(fmt.Stringer).String()),
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

// PatternFromURL convert relative URL to pattern by necessary operations, such as remove #fragment portion
func PatternFromURL(relativeUrl string) string {
	u, e := url.Parse(relativeUrl)
	if e != nil {
		split := strings.SplitN(relativeUrl, "#", 2)
		return split[0]
	}
	return u.Path
}

/**************************
	helpers
***************************/
func interfaceToRoute(i interface{}) (*web.Route, error) {
	switch v := i.(type) {
	case web.Route:
		return &v, nil
	case *web.Route:
		return v, nil
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
	var desc string
	if stringer, ok := m.(fmt.Stringer); ok {
		desc = stringer.String()
	}
	return &routeMatcher{
		description: desc,
		delegate: m,
	}
}