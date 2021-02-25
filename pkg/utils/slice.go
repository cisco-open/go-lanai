package utils

import "strings"

type CommaSeparatedSlice []string

// fmt.Stringer
func (s CommaSeparatedSlice) String() string {
	return strings.Join(s, ", ")
}

// encoding.TextMarshaler
func (s CommaSeparatedSlice) MarshalText() ([]byte, error) {
	return []byte(s.String()), nil
}

// encoding.TextUnmarshaler
func (s *CommaSeparatedSlice) UnmarshalText(data []byte) error {
	result := []string{}
	split := strings.Split(string(data), ",")
	for _, s := range split {
		s = strings.TrimSpace(s)
		result = append(result, s)
	}
	*s = result
	return nil
}


