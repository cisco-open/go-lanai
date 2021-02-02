package idp

import (
	"context"
	util_matcher "cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/matcher"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/matcher"
	"net"
	"net/http"
)

const (
	InternalIdpForm = "InternalIdpForm"
	ExternalIdpSAML = "ExternalIdpSAML"
)

type AuthFlowManager interface {
	GetAuthFlow(domain string) (string, error)
}

func RequestWithAuthenticationMethod(authMethod string, authFlowManager AuthFlowManager) web.RequestMatcher {
	matchable := func(ctx context.Context, request *http.Request) (interface{}, error) {
		host, _, err := net.SplitHostPort(request.Host)
		if err != nil {
			host = request.Host
		}
		authMethod, err := authFlowManager.GetAuthFlow(host)

		if err == nil {
			return authMethod, nil
		} else {
			return InternalIdpForm, nil
		}
	}
	return matcher.CustomMatcher("authentication flow matcher", matchable , util_matcher.WithString(authMethod, true))
}


