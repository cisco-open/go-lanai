package log

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log/internal"
)

func IsTerminal(l Logger) bool {
	logger, ok := l.(*configurableLogger)
	return ok && logger.isTerm
}

// Capped truncate given value to specified length, with tailing "..." if truncated
func Capped(v interface{}, cap int) string {
	return internal.Capped(v, cap)
}

// Padding example: `Padding("some string", -20)`
func Padding(v interface{}, padding int) string {
	return internal.Padding(v, padding)
}

func DebugShowcase() {
	internal.DebugShowcase()
}
