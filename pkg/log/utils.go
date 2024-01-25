package log

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log/internal"
)

func IsTerminal(l Logger) bool {
	v, ok := l.(internal.TerminalAware)
	return ok && v.IsTerminal()
}

// Capped truncate given value to specified length
// if cap > 0: with tailing "..." if truncated
// if cap < 0: with middle "..." if truncated
func Capped(v interface{}, cap int) string {
	return internal.Capped(cap, v)
}

// Padding example: `Padding("some string", -20)`
func Padding(v interface{}, padding int) string {
	return internal.Padding(padding, v)
}

func DebugShowcase() {
	internal.DebugShowcase()
}
