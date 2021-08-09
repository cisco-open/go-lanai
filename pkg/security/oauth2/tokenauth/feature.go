package tokenauth

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"fmt"
)

//goland:noinspection GoNameStartsWithPackageName
type TokenAuthFeature struct {
	errorHandler    *OAuth2ErrorHandler
	postBodyEnabled bool
}

func (f *TokenAuthFeature) Identifier() security.FeatureIdentifier {
	return FeatureId
}

// Configure Standard security.Feature entrypoint
// use (*access.AccessControl).AllowIf(ScopesApproved(...)) for scope based access decision maker
func Configure(ws security.WebSecurity) *TokenAuthFeature {
	feature := New()
	if fc, ok := ws.(security.FeatureModifier); ok {
		return fc.Enable(feature).(*TokenAuthFeature)
	}
	panic(fmt.Errorf("unable to configure oauth2 authserver: provided WebSecurity [%T] doesn't support FeatureModifier", ws))
}

// New Standard security.Feature entrypoint, DSL style. Used with security.WebSecurity
// use (*access.AccessControl).AllowIf(ScopesApproved(...)) for scope based access decision maker
func New() *TokenAuthFeature {
	return &TokenAuthFeature{}
}

/** Setters **/

func (f *TokenAuthFeature) ErrorHandler(errorHandler *OAuth2ErrorHandler) *TokenAuthFeature {
	f.errorHandler = errorHandler
	return f
}

func (f *TokenAuthFeature) EnablePostBody() *TokenAuthFeature {
	f.postBodyEnabled = true
	return f
}
