package web

const (
	LowestMiddlewareOrder = int(^uint(0) >> 1) // max int
	HighestMiddlewareOrder = -LowestMiddlewareOrder - 1 // min int

	ContextKeySecurity = "Security"
	ContextKeySession = "Session"
	ContextKeyContextPath = "ContextPath"
)
