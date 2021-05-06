package utils

import (
	"encoding/json"
	"strings"
)

type CommaSeparatedSlice []string

// fmt.Stringer
func (s CommaSeparatedSlice) String() string {
	return strings.Join(s, ", ")
}

// MarshalText encoding.TextMarshaler
func (s CommaSeparatedSlice) MarshalText() ([]byte, error) {
	return []byte(s.String()), nil
}

// UnmarshalText encoding.TextUnmarshaler
func (s *CommaSeparatedSlice) UnmarshalText(data []byte) error {
	var result []string
	split := strings.Split(string(data), ",")
	for _, s := range split {
		s = strings.TrimSpace(s)
		result = append(result, s)
	}
	*s = result
	return nil
}

// UnmarshalJSON json.Unmarshaler
func (s *CommaSeparatedSlice) UnmarshalJSON(data []byte) error {
	// try comma separated format
	var str string
	if e := json.Unmarshal(data, &str); e == nil {
		return s.UnmarshalText([]byte(str))
	}

	// try regular array
	var result []string
	if e := json.Unmarshal(data, &result); e == nil {
		return nil
	}
	*s = result
	return nil
}


