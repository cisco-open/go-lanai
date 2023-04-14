package schematotext

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/cmd/lanai-cli/codegen/generator/template_funcs/go"
	"cto-github.cisco.com/NFV-BU/go-lanai/cmd/lanai-cli/codegen/generator/template_funcs/lanai/lanaiutil"
	"fmt"
	"github.com/getkin/kin-openapi/openapi3"
	"path"
	"strings"
)

type objectType struct {
	data                  interface{}
	defaultObjectName     string
	currentPkg            string
	restrictExternalTypes bool
}

func NewObjectType(data interface{}, opts ...func(option *translatorOptions)) *objectType {
	o := &translatorOptions{}
	for _, fn := range opts {
		fn(o)
	}

	return &objectType{
		data:                  data,
		defaultObjectName:     o.defaultObjectName,
		currentPkg:            o.currentPkg,
		restrictExternalTypes: o.restrictExternalTypes,
	}
}

func (o objectType) ToText() (result string, err error) {
	schema, err := lanaiutil.ConvertToSchemaRef(o.data)
	if err != nil {
		return "", err
	}
	switch schema.Value.Type {
	case openapi3.TypeObject:
		if len(schema.Value.Properties) == 0 {
			if schema.Value.AdditionalPropertiesAllowed != nil && *schema.Value.AdditionalPropertiesAllowed {
				result = "map[string]interface{}"
			} else {
				additionalPropertyType, err := NewDataTypeToTextTranslator(
					schema.Value.AdditionalProperties,
					WithCurrentPackage(o.currentPkg),
					WithRestrictExternalTypes(o.restrictExternalTypes),
					WithDefaultObjectName("interface{}")).ToText()
				if err != nil {
					return "", err
				}
				result = "map[string]" + additionalPropertyType
			}
			return result, nil
		}
		fallthrough
	default:
		if schema.Ref == "" {
			result = o.defaultObjectName
		} else {
			result = path.Base(schema.Ref)
			refPackage, ok := _go.StructRegistry()[strings.ToLower(result)]
			if ok && refPackage != o.currentPkg {
				result = fmt.Sprintf("%v.%v", path.Base(refPackage), result)
			}
		}
		return result, nil
	}
}
