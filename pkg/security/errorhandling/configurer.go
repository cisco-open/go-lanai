package errorhandling

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/middleware"
	"fmt"
)

const (
	FeatureId       = "ErrorHandling"
)

// We currently don't have any stuff to configure
//goland:noinspection GoNameStartsWithPackageName
type ErrorHandlingFeature struct {
	authEntryPoint security.AuthenticationEntryPoint
	accessDeniedHandler security.AccessDeniedHandler
	authErrorHandler security.AuthenticationErrorHandler
}

// Standard security.Feature entrypoint
func (f *ErrorHandlingFeature) Identifier() security.FeatureIdentifier {
	return FeatureId
}

func (f *ErrorHandlingFeature) AuthenticationEntryPoint(v security.AuthenticationEntryPoint) *ErrorHandlingFeature {
	f.authEntryPoint = v
	return f
}

func (f *ErrorHandlingFeature) AccessDeniedHandler(v security.AccessDeniedHandler) *ErrorHandlingFeature {
	f.accessDeniedHandler = v
	return f
}

// AuthenticationErrorHandler set authentication error handler override.
//If AuthenticationEntryPoint is provided, this value is ignored
func (f *ErrorHandlingFeature) AuthenticationErrorHandler(v security.AuthenticationErrorHandler) *ErrorHandlingFeature {
	f.authErrorHandler = v
	return f
}

func Configure(ws security.WebSecurity) *ErrorHandlingFeature {
	feature := New()
	if fc, ok := ws.(security.FeatureModifier); ok {
		_ = fc.Enable(feature) // we ignore error here
		return feature
	}
	panic(fmt.Errorf("unable to configure session: provided WebSecurity [%T] doesn't support FeatureModifier", ws))
}

// Standard security.Feature entrypoint, DSL style. Used with security.WebSecurity
func New() *ErrorHandlingFeature {
	return &ErrorHandlingFeature{}
}

//goland:noinspection GoNameStartsWithPackageName
type ErrorHandlingConfigurer struct {

}

func newErrorHandlingConfigurer() *ErrorHandlingConfigurer {
	return &ErrorHandlingConfigurer{
	}
}

func (ehc *ErrorHandlingConfigurer) Apply(feature security.Feature, ws security.WebSecurity) error {
	// Validate
	if err := ehc.validate(feature.(*ErrorHandlingFeature), ws); err != nil {
		return err
	}
	f := feature.(*ErrorHandlingFeature)

	mw := NewErrorHandlingMiddleware()
	mw.accessDeniedHandler = f.accessDeniedHandler
	mw.entryPoint = f.authEntryPoint
	mw.authErrorHandler = f.authErrorHandler

	errHandler := middleware.NewBuilder("error handling").
		Order(security.MWOrderErrorHandling).
		Use(mw.HandlerFunc())

	ws.Add(errHandler)
	return nil
}


func (ehc *ErrorHandlingConfigurer) validate(f *ErrorHandlingFeature, ws security.WebSecurity) error {
	if f.authEntryPoint != nil && f.authErrorHandler != nil {
		fmt.Printf("for route matches [%v], authentication error handler will be ignored because entry point is set", ws)
	}

	if f.authEntryPoint == nil && f.authErrorHandler == nil {
		fmt.Printf("for route matches [%v], using default authentication error handler", ws)
		f.authErrorHandler = &security.DefaultAuthenticationErrorHandler{}
	}

	if f.accessDeniedHandler == nil {
		fmt.Printf("for route matches [%v], using default access denied handler", ws)
		f.accessDeniedHandler = &security.DefaultAccessDeniedHandler{}
	}
	return nil
}