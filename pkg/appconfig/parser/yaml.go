package parser

import (
	//"gopkg.in/yaml.v2"
	"encoding/json"
	"github.com/ghodss/yaml"
)

func NewYamlPropertyParser() PropertyParser {

	return func(encoded []byte) (map[string]interface{}, error){
		m := make(map[string]interface{})
		encodedJson, e := yaml.YAMLToJSON(encoded) //need to do this because json marshal needs to work on map with string key. so only json marshal and unmarshal matches
		if e != nil {
			return m, e
		}

		e = json.Unmarshal(encodedJson, &m)
		return m, e
	}
}
