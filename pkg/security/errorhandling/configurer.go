package errorhandling

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/middleware"
	"fmt"
)

var (
	FeatureId       = security.FeatureId("ErrorHandling", security.FeatureOrderErrorHandling)
)

// We currently don't have any stuff to configure
//goland:noinspection GoNameStartsWithPackageName
type ErrorHandlingFeature struct {
	authEntryPoint      security.AuthenticationEntryPoint
	accessDeniedHandler security.AccessDeniedHandler
	authErrorHandler    security.AuthenticationErrorHandler
	errorHandler        *security.CompositeErrorHandler
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

func (f *ErrorHandlingFeature) AuthenticationErrorHandler(v security.AuthenticationErrorHandler) *ErrorHandlingFeature {
	f.authErrorHandler = v
	return f
}

// AdditionalErrorHandler add security.ErrorHandler to existing list.
// This value is typically used by other features, because there are no other type of concrete errors except for
// AuthenticationError and AccessControlError,
// which are handled by AccessDeniedHandler, AuthenticationErrorHandler and AuthenticationEntryPoint
func (f *ErrorHandlingFeature) AdditionalErrorHandler(v security.ErrorHandler) *ErrorHandlingFeature {
	f.errorHandler.Add(v)
	return f
}

func Configure(ws security.WebSecurity) *ErrorHandlingFeature {
	feature := New()
	if fc, ok := ws.(security.FeatureModifier); ok {
		return fc.Enable(feature).(*ErrorHandlingFeature)
	}
	panic(fmt.Errorf("unable to configure session: provided WebSecurity [%T] doesn't support FeatureModifier", ws))
}

// Standard security.Feature entrypoint, DSL style. Used with security.WebSecurity
func New() *ErrorHandlingFeature {
	return &ErrorHandlingFeature{
		errorHandler: security.NewErrorHandler(),
	}
}

//goland:noinspection GoNameStartsWithPackageName
type ErrorHandlingConfigurer struct {

}

func newErrorHandlingConfigurer() *ErrorHandlingConfigurer {
	return &ErrorHandlingConfigurer{
	}
}

func (ehc *ErrorHandlingConfigurer) Apply(feature security.Feature, ws security.WebSecurity) error {
	// Verify
	if err := ehc.validate(feature.(*ErrorHandlingFeature), ws); err != nil {
		return err
	}
	f := feature.(*ErrorHandlingFeature)

	authErrorHandler := ws.Shared(security.WSSharedKeyCompositeAuthErrorHandler).(*security.CompositeAuthenticationErrorHandler)
	authErrorHandler.Add(f.authErrorHandler)

	accessDeniedHandler := ws.Shared(security.WSSharedKeyCompositeAccessDeniedHandler).(*security.CompositeAccessDeniedHandler)
	accessDeniedHandler.Add(f.accessDeniedHandler)

	mw := NewErrorHandlingMiddleware()
	mw.entryPoint = f.authEntryPoint
	mw.accessDeniedHandler = accessDeniedHandler
	mw.authErrorHandler = authErrorHandler
	mw.errorHandler = f.errorHandler

	errHandler := middleware.NewBuilder("error handling").
		Order(security.MWOrderErrorHandling).
		Use(mw.HandlerFunc())

	ws.Add(errHandler)
	return nil
}


func (ehc *ErrorHandlingConfigurer) validate(f *ErrorHandlingFeature, ws security.WebSecurity) error {
	if f.authEntryPoint == nil {
		logger.Infof("authentication entry point is not set, fallback to access denied handler - [%v], ", log.Capped(ws, 80))
	}

	if f.authErrorHandler == nil {
		logger.Infof("using default authentication error handler - [%v]", log.Capped(ws, 80))
		f.authErrorHandler = &security.DefaultAuthenticationErrorHandler{}
	}

	if f.accessDeniedHandler == nil {
		logger.Infof("using default access denied handler - [%v]", log.Capped(ws, 80))
		f.accessDeniedHandler = &security.DefaultAccessDeniedHandler{}
	}

	// always add a default to the end. note: DefaultErrorHandler is unordered
	f.errorHandler.Add(&security.DefaultErrorHandler{})
	return nil
}

