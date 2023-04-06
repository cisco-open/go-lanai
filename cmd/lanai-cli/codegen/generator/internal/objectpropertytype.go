package internal

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"encoding/json"
	"fmt"
	"github.com/getkin/kin-openapi/openapi3"
	"path"
	"reflect"
	"strings"
)

type ObjectType struct {
	data                  interface{}
	defaultObjectName     string
	currentPkg            string
	restrictExternalTypes bool
}

func NewObjectType(data interface{}, opts ...func(option *TranslatorOptions)) *ObjectType {
	o := &TranslatorOptions{}
	for _, fn := range opts {
		fn(o)
	}

	return &ObjectType{
		data:                  data,
		defaultObjectName:     o.defaultObjectName,
		currentPkg:            o.currentPkg,
		restrictExternalTypes: o.restrictExternalTypes,
	}
}

func (o ObjectType) toText() (result string, err error) {
	schema, err := convertToSchemaRef(o.data)
	if err != nil {
		return "", err
	}
	switch schema.Value.Type {
	case openapi3.TypeObject:
		if len(schema.Value.Properties) == 0 {
			if utils.FromPtr(schema.Value.AdditionalPropertiesAllowed) {
				result = reflect.TypeOf(json.RawMessage{}).String()
			} else {
				additionalPropertyType, err := getDataTypeToTextTranslator(
					schema.Value.AdditionalProperties,
					WithCurrentPackage(o.currentPkg),
					WithRestrictExternalTypes(o.restrictExternalTypes),
					WithDefaultObjectName("")).toText()
				if err != nil {
					return "", err
				}

				if additionalPropertyType == "" {
					result = reflect.TypeOf(json.RawMessage{}).String()
				} else {
					result = "map[string]" + additionalPropertyType
				}
			}
			return result, nil
		}
		fallthrough
	default:
		if schema.Ref == "" {
			result = o.defaultObjectName
		} else {
			result = path.Base(schema.Ref)
			refPackage, ok := structRegistry[strings.ToLower(result)]
			if ok && refPackage != o.currentPkg {
				result = fmt.Sprintf("%v.%v", path.Base(refPackage), result)
			}
		}
		return result, nil
	}
}
