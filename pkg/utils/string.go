package utils

import "strconv"

func UnQuote(s string) string {
	if len(s) >= 2 {
		if c := s[len(s)-1]; s[0] == c && (c == '"' || c == '\'') {
			return s[1 : len(s)-1]
		}
	}
	return s
}

func ParseString(s string) interface{} {
	//try number before boolean because 1/0 can be parsed to boolean true/false
	if numVal, err := strconv.ParseFloat(s, 64); err == nil {
		return numVal
	} else if boolVal, err := strconv.ParseBool(s); err == nil {
		return boolVal
	} else {
		return s
	}
}