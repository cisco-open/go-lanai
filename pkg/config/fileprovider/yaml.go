package fileprovider

import (
	"encoding/json"
	"github.com/ghodss/yaml" //TODO: revisit library usage
)

func NewYamlPropertyParser() PropertyParser {

	return func (reader ContentReader) (map[string]interface{}, error){
		encodedYAML, err := reader()
		if err != nil {
			return nil, err
		}

		//TODO: revisit here to see if there's approach without having to convert to json

		encodedJSON, err := yaml.YAMLToJSON(encodedYAML)
		if err != nil {
			return nil, err
		}

		decodedJSON := map[string]interface{}{}
		if err := json.Unmarshal(encodedJSON, &decodedJSON); err != nil {
			return nil, err
		}

		return FlattenJSON(decodedJSON, "")
	}
}