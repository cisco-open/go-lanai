package authorize

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/auth"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"fmt"
)

// AuthorizeFeature configures authorization endpoints
//goland:noinspection GoNameStartsWithPackageName
type AuthorizeFeature struct {
	path             string
	condition        web.RequestMatcher
	approvalPath     string
	requestProcessor auth.AuthorizeRequestProcessor
	authorizeHandler auth.AuthorizeHandler
	errorHandler     *auth.OAuth2ErrorHandler
}

func (f *AuthorizeFeature) Identifier() security.FeatureIdentifier {
	return FeatureId
}

// Configure is standard security.Feature entrypoint
func Configure(ws security.WebSecurity) *AuthorizeFeature {
	feature := NewEndpoint()
	if fc, ok := ws.(security.FeatureModifier); ok {
		return fc.Enable(feature).(*AuthorizeFeature)
	}
	panic(fmt.Errorf("unable to configure oauth2 authserver: provided WebSecurity [%T] doesn't support FeatureModifier", ws))
}

// NewEndpoint is standard security.Feature entrypoint, DSL style. Used with security.WebSecurity
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

func (f *AuthorizeFeature) ApprovalPath(approvalPath string) *AuthorizeFeature {
	f.approvalPath = approvalPath
	return f
}

func (f *AuthorizeFeature) RequestProcessors(processors ...auth.ChainedAuthorizeRequestProcessor) *AuthorizeFeature {
	f.requestProcessor = auth.NewAuthorizeRequestProcessor(processors...)
	return f
}

func (f *AuthorizeFeature) RequestProcessor(processor auth.AuthorizeRequestProcessor) *AuthorizeFeature {
	f.requestProcessor = processor
	return f
}

func (f *AuthorizeFeature) ErrorHandler(errorHandler *auth.OAuth2ErrorHandler) *AuthorizeFeature {
	f.errorHandler = errorHandler
	return f
}

func (f *AuthorizeFeature) AuthorizeHanlder(authHanlder auth.AuthorizeHandler) *AuthorizeFeature {
	f.authorizeHandler = authHanlder
	return f
}