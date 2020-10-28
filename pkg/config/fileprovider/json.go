package fileprovider

import (
	"encoding/json"
)

//import (
//	"context"
//	"cto-github.cisco.com/livdu/jupiter/pkg/config"
//	"fmt"
//	"github.com/pkg/errors"
//	json "github.com/yosuke-furukawa/json5/encoding/json5" //TODO: revisit library usage
//	"io"
//)
//
////TODO: different package
//
//type JSONFile struct {
//	name   string
//	path   string
//}
//
//func (f *JSONFile) Description() string {
//	return fmt.Sprintf("%s: [%s]", f.name, f.path)
//}
//
//func (f *JSONFile) Load(ctx context.Context) (map[string]interface{}, error) {
//	fmt.Printf("Loading JSON config: %s", f.path)
//
//	encodedJSON, err := f.reader()
//	if err != nil {
//		return nil, err
//	}
//
//	decodedJSON := map[string]interface{}{}
//	if err := json.Unmarshal(encodedJSON, &decodedJSON); err != nil {
//		return nil, err
//	}
//
//	settings, err := FlattenJSON(decodedJSON, "")
//	if err != nil {
//		return nil, err
//	}
//
//	if ctx.Err() != nil {
//		return nil, errors.Wrap(ctx.Err(), "Failed to load JSON config")
//	}
//
//	return settings, nil
//
//}
//
//func FlattenJSON(input map[string]interface{}, namespace string) (map[string]interface{}, error) {
//	return flattenMap(input, namespace)
//}
//
//func flattenMap(input map[string]interface{}, namespace string) (map[string]interface{}, error) {
//	flattened := make(map[string]interface{})
//
//	for key, value := range input {
//		var token string
//		if namespace == "" {
//			token = key
//		} else {
//			token = fmt.Sprintf("%s.%s", namespace, key)
//		}
//
//		settings, err := flattenChild(value, token)
//		if err != nil {
//			return nil, err
//		}
//
//		for k, v := range settings {
//			flattened[config.NormalizeKey(k)] = v
//		}
//	}
//
//	return flattened, nil
//}
//
//func flattenChild(value interface{}, token string) (map[string]interface{}, error) {
//	if objectChild, ok := value.(map[string]interface{}); ok {
//		return flattenMap(objectChild, token)
//	} else if arrayChild, ok := value.([]interface{}); ok {
//		return flattenArray(arrayChild, token)
//	} else { // Scalar
//		key := config.NormalizeKey(token)
//		value := fmt.Sprintf("%v", value)
//		return map[string]interface{}{key: value}, nil
//	}
//}
//
//func flattenArray(input []interface{}, namespace string) (map[string]interface{}, error) {
//	flattened := make(map[string]interface{})
//	for k, v := range input {
//		token := config.NormalizeKey(namespace) + fmt.Sprintf("[%d]", k)
//
//		settings, err := flattenChild(v, token)
//		if err != nil {
//			return nil, err
//		}
//
//		for k, v := range settings {
//			flattened[config.NormalizeKey(k)] = v
//		}
//	}
//	return flattened, nil
//}
//
//func NewJSONFile(name, path string, reader ContentReader) *JSONFile {
//	return &JSONFile{
//		name:   name,
//		path:   path,
//		reader: reader,
//	}
//}


func NewJSONPropertyParser() PropertyParser {
	return func(encoded []byte) (map[string]interface{}, error) {
		var m = make(map[string]interface{})
		error := json.Unmarshal(encoded, &m)
		return m, error
	}
}