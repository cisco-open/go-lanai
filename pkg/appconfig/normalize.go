package appconfig

import (
	"strings"
)

func NormalizeKey(key string) string {
	return strings.ToLower(
		strings.ReplaceAll(
			strings.ReplaceAll(
				key,
				"-",
				""),
			"_",
			"."))
}