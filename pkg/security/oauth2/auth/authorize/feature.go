package authorize

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/auth"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"fmt"
)

// We currently don't have any stuff to configure
//goland:noinspection GoNameStartsWithPackageName
type AuthorizeFeature struct {
	path             string
	condition 		 web.RequestMatcher
	requestProcessor *auth.CompositeAuthorizeRequestProcessor
	errorHandler     *auth.OAuth2ErrorHandler
}

// Standard security.Feature entrypoint
func (f *AuthorizeFeature) Identifier() security.FeatureIdentifier {
	return FeatureId
}

func Configure(ws security.WebSecurity) *AuthorizeFeature {
	feature := NewEndpoint()
	if fc, ok := ws.(security.FeatureModifier); ok {
		return fc.Enable(feature).(*AuthorizeFeature)
	}
	panic(fmt.Errorf("unable to configure oauth2 authserver: provided WebSecurity [%T] doesn't support FeatureModifier", ws))
}

// Standard security.Feature entrypoint, DSL style. Used with security.WebSecurity
func NewEndpoint() *AuthorizeFeature {
	return &AuthorizeFeature{
	}
}

/** Setters **/
func (f *AuthorizeFeature) Path(path string) *AuthorizeFeature {
	f.path = path
	return f
}

func (f *AuthorizeFeature) Condition(condition web.RequestMatcher) *AuthorizeFeature {
	f.condition = condition
	return f
}

func (f *AuthorizeFeature) RequestProcessors(processors ...auth.AuthorizeRequestProcessor) *AuthorizeFeature {
	f.requestProcessor = auth.NewCompositeAuthorizeRequestProcessor(processors...)
	return f
}

func (f *AuthorizeFeature) ErrorHandler(errorHandler *auth.OAuth2ErrorHandler) *AuthorizeFeature {
	f.errorHandler = errorHandler
	return f
}