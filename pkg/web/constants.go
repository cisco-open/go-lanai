package web

type EmptyRequest struct {}

const (
	LowestMiddlewareOrder  = int(^uint(0) >> 1)         // max int
	HighestMiddlewareOrder = -LowestMiddlewareOrder - 1 // min int

	ErrorTemplate = "error.tmpl"

	ContextKeySecurity    = "Security"
	ContextKeySession     = "Session"
	ContextKeyContextPath = "ContextPath"
	ContextKeyCsrf        = "CSRF"

	MethodAny = "ANY"
)

const (
	HeaderAuthorization      = "Authorization"
	HeaderOrigin             = "Origin"
	HeaderACAllowCredentials = "Access-Control-Allow-Credentials"
	HeaderACAllowHeaders     = "Access-Control-Allow-Headers"
	HeaderACAllowMethods     = "Access-Control-Allow-AllowedMethodsStr"
	HeaderACAllowOrigin      = "Access-Control-Allow-Origin"
	HeaderACExposeHeaders    = "Access-Control-Expose-Headers"
	HeaderACMaxAge           = "Access-Control-Max-Age"
	HeaderACRequestHeaders   = "Access-Control-Request-Headers"
	HeaderACRequestMethod    = "Access-Control-Request-Method"
)