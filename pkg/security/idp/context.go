package idp

import (
	"context"
	util_matcher "cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/matcher"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/matcher"
	"fmt"
	"net"
	"net/http"
)

const (
	InternalIdpForm = AuthenticationFlow("InternalIdpForm")
	ExternalIdpSAML = AuthenticationFlow("ExternalIdpSAML")
	UnknownIdp      = AuthenticationFlow("UnKnown")
)

type AuthenticationFlow string

// encoding.TextMarshaler
func (f AuthenticationFlow) MarshalText() ([]byte, error) {
	return []byte(f), nil
}

// encoding.TextUnmarshaler
func (f *AuthenticationFlow) UnmarshalText(data []byte) error {
	value := string(data)
	switch value {
	case string(InternalIdpForm):
		*f = InternalIdpForm
	case string(ExternalIdpSAML):
		*f = ExternalIdpSAML
	default:
		return fmt.Errorf("unrecognized authentication flow: %s", value)
	}
	return nil
}

type IdentityProvider interface {
	Domain() string
}

type AuthenticationFlowAware interface {
	AuthenticationFlow() AuthenticationFlow
}

type IdentityProviderManager interface {
	GetIdentityProvidersWithFlow(flow AuthenticationFlow) []IdentityProvider
	GetIdentityProviderByDomain(domain string) (IdentityProvider, error)
}

func RequestWithAuthenticationFlow(flow AuthenticationFlow, idpManager IdentityProviderManager) web.RequestMatcher {
	matchableError := func() (interface{}, error) {
		return string(UnknownIdp), nil
	}

	matchable := func(_ context.Context, request *http.Request) (interface{}, error) {
		host, _, err := net.SplitHostPort(request.Host)
		if err != nil {
			host = request.Host
		}

		idp, err := idpManager.GetIdentityProviderByDomain(host)
		if err != nil {
			return matchableError()
		}

		fa, ok := idp.(AuthenticationFlowAware)
		if !ok {
			return matchableError()
		}
		return string(fa.AuthenticationFlow()), nil
	}

	return matcher.CustomMatcher(fmt.Sprintf("IDP with [%s]", flow),
		matchable,
		util_matcher.WithString(string(flow), true))
}