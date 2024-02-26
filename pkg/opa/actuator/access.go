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

package opaactuator

import (
	"github.com/cisco-open/go-lanai/pkg/actuator"
	"github.com/cisco-open/go-lanai/pkg/opa"
	opaaccess "github.com/cisco-open/go-lanai/pkg/opa/access"
	"github.com/cisco-open/go-lanai/pkg/security/access"
	"github.com/cisco-open/go-lanai/pkg/web"
	"github.com/cisco-open/go-lanai/pkg/web/matcher"
	"regexp"
)

const RequestInputKeyEndpointID = `endpoint_id`

func NewAccessControlWithOPA(props actuator.SecurityProperties, opts ...opa.RequestQueryOptions) actuator.AccessControlCustomizer {
	return actuator.AccessControlCustomizeFunc(func(ac *access.AccessControlFeature, epId string, paths []string) {
		if len(paths) == 0 {
			return
		}

		// configure request matchers
		reqMatcher := pathToRequestPattern(paths[0])
		for _, p := range paths[1:] {
			reqMatcher = reqMatcher.Or(pathToRequestPattern(p))
		}

		switch {
		case !isSecurityEnabled(epId, &props):
			ac.Request(reqMatcher).PermitAll()
		default:
			opts = append(opts, func(opt *opa.RequestQuery) {
				opt.ExtraData[RequestInputKeyEndpointID] = epId
			})
			ac.Request(reqMatcher).CustomDecisionMaker(opaaccess.DecisionMakerWithOPA(opts...))
		}
	})
}

var pathVarRegex = regexp.MustCompile(`:[a-zA-Z0-9\-_]+`)

// pathToRequestPattern convert path variables to wildcard request pattern
// "/path/to/:any/endpoint" would converted to "/path/to/*/endpoint
func pathToRequestPattern(path string) web.RequestMatcher {
	patternStr := pathVarRegex.ReplaceAllString(path, "*")
	return matcher.RequestWithPattern(patternStr)
}

func isSecurityEnabled(epId string, properties *actuator.SecurityProperties) bool {
	enabled := properties.EnabledByDefault
	if props, ok := properties.Endpoints[epId]; ok {
		if props.Enabled != nil {
			enabled = *props.Enabled
		}
	}
	return enabled
}
