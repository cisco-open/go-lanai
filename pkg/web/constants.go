package web

const (
	LowestMiddlewareOrder = int(^uint(0) >> 1) // max int
	HighestMiddlewareOrder = -LowestMiddlewareOrder - 1 // min int

	ErrorTemplate = "error.tmpl"

	ContextKeySecurity = "Security"
	ContextKeySession = "Session"
	ContextKeyContextPath = "ContextPath"
)

type EmptyRequest struct {}