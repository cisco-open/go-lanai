package parser

import (
	"encoding/json"
)

func NewJSONPropertyParser() PropertyParser {
	return func(encoded []byte) (map[string]interface{}, error) {
		var m = make(map[string]interface{})
		error := json.Unmarshal(encoded, &m)
		return m, error
	}
}