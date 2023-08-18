package opaactuator

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/actuator"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/opa"
	opaaccess "cto-github.cisco.com/NFV-BU/go-lanai/pkg/opa/access"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/access"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/matcher"
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
