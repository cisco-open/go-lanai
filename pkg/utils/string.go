package utils

import (
	"encoding/json"
	"strconv"
	"strings"
)

func UnQuote(s string) string {
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
