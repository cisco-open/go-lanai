package utils

import (
	"encoding/json"
	"strconv"
	"strings"
	"unicode"
)

func UnQuote(s string) string {
	s = strings.TrimSpace(s)
	if len(s) >= 2 {
		if c := s[len(s)-1]; s[0] == c && (c == '"' || c == '\'') {
			return s[1 : len(s)-1]
		}
	}
	return s
}

func ParseString(s string) interface{} {
	// try number before boolean because 1/0 can be parsed to boolean true/false
	if numVal, err := strconv.ParseFloat(s, 64); err == nil {
		return numVal
	} else if boolVal, err := strconv.ParseBool(s); err == nil {
		return boolVal
	}

	// we also support []interface{} and map[string]interface{}
	trimmed := strings.TrimSpace(s)
	if strings.HasPrefix(trimmed, "{") {
		var v map[string]interface{}
		if e := json.Unmarshal([]byte(trimmed), &v); e == nil {
			return v
		}
	}

	if strings.HasPrefix(trimmed, "[") {
		var v []interface{}
		if e := json.Unmarshal([]byte(trimmed), &v); e == nil {
			return v
		}
	}

	return s
}

const dash = rune('-')

// CamelToSnakeCase convert "camelCase" string to "snake-case"
func CamelToSnakeCase(camelCase string) string {
	var converted []rune
	for pos, char := range camelCase {
		if unicode.IsUpper(char) {
			if pos>0 && unicode.IsLower([]rune(camelCase)[pos-1]) {
				converted = append(converted, dash)
			}
			converted = append(converted, unicode.ToLower(char))
		} else {
			converted = append(converted, unicode.ToLower(char))
		}
	}
	return string(converted)
}
