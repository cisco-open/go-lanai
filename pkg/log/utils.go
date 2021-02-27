package log

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log/internal"
	"fmt"
	"strconv"
)

func Capped(v interface{}, cap int) string {
	return CappedString(internal.Sprint(v), cap)
}

func CappedString(s string, cap int) string {
	if len(s) <= cap {
		return s
	}
	return fmt.Sprintf("%." + strconv.Itoa(cap - 3) + "s...", s)
}

