package internal

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/cmd/lanai-cli/codegen/generator/internal/representation"
	"fmt"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/google/uuid"
	"path"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"text/template"
)

var (
	propertyFuncMap = template.FuncMap{
		"property": NewProperty,
	}
)

type Property struct {
	PropertyName string
	TypePrefix   string
	OmitJSON     bool
	RequiredList []string
	PropertyData interface{}
}

func NewProperty(data interface{}, name string, requiredList []string, typePrefix ...string) Property {
	return Property{
		PropertyName: name,
		PropertyData: data,
		RequiredList: requiredList,
		TypePrefix:   strings.Join(typePrefix, ""),
	}
}

func (p Property) SetOmitJSON(val bool) Property {
	p.OmitJSON = val
	return p
}

func propertyToGoType(p Property, currentPkg string) (string, error) {
	defaultObjectName := toTitle(p.TypePrefix + toTitle(p.PropertyName))
	result, err := schemaToText(p.PropertyData, defaultObjectName, currentPkg)
	if err != nil {
		return "", err
	}
	return result, nil
}

func schemaToText(element interface{}, defaultObjectName string, currentPkg string) (result string, err error) {
	schema, err := convertToSchemaRef(element)
	if err != nil {
		return "", err
	}
	// for the non-object things
	switch schema.Value.Type {
	case openapi3.TypeBoolean:
		result = reflect.TypeOf(true).String()
	case openapi3.TypeNumber:
		result = reflect.TypeOf(1.1).String()
	case openapi3.TypeInteger:
		result = reflect.TypeOf(1).String()
	case openapi3.TypeString:
		if shouldBeUUIDType(element) {
			result = reflect.TypeOf(uuid.UUID{}).String()
		} else {
			result = reflect.TypeOf("string").String()
		}
	case openapi3.TypeArray:
		inner, err := schemaToText(schema.Value.Items, defaultObjectName, currentPkg)
		if err != nil {
			return "", err
		}
		result = "[]" + inner
	case openapi3.TypeObject:
		if len(schema.Value.Properties) == 0 {
			s := representation.NewSchema("", schema)
			if s.HasAdditionalProperties() {
				innerType := ""
				if schema.Value.AdditionalPropertiesAllowed != nil && *schema.Value.AdditionalPropertiesAllowed {
					innerType = "interface{}"
				} else {
					innerType, err = schemaToText(schema.Value.AdditionalProperties, "interface{}", currentPkg)
					if err != nil {
						return "", err
					}
				}
				result = "map[string]" + innerType
				return result, nil
			}
		}
		fallthrough
	default:
		if schema.Ref != "" {
			result = path.Base(schema.Ref)
			refPackage, ok := structRegistry[strings.ToLower(result)]
			if ok && refPackage != currentPkg {
				result = fmt.Sprintf("%v.%v", path.Base(refPackage), result)
			}
		} else {
			result = defaultObjectName
		}
	}

	return result, nil
}

func shouldHavePointer(p Property) bool {
	isRequired := listContains(p.RequiredList, p.PropertyName)
	schema, _ := convertToSchemaRef(p.PropertyData)
	if schema.Value.Enum != nil {
		return false
	}
	if len(schema.Value.Properties) == 0 {
		s := representation.NewSchema("", schema)
		if s.HasAdditionalProperties() {
			return false
		}
	}
	if schema.Value.Nullable {
		return true
	}
	if schema.Value.Type == "object" && isRequired {
		return true
	}
	return valuePassesValidation(schema.Value, zeroValue(schema.Value))
}

func valuePassesValidation(schema *openapi3.Schema, value reflect.Value) (result bool) {
	switch value.Kind() {
	case reflect.String:
		if rValue, _ := regex(*schema); rValue != nil {
			found, err := regexp.MatchString(rValue.Value, value.String())
			if err == nil || !found {
				return false
			}
		}
		fallthrough
	default:
		return valueIsWithinSchemaLimits(schema, value)
	}
}

func valueIsWithinSchemaLimits(schema *openapi3.Schema, value reflect.Value) bool {
	min, max := limitsForSchema(schema)
	switch value.Kind() {
	case reflect.String:
		return !isOutOfBounds(len(value.String()), min, max)
	case reflect.Int:
		return !isOutOfBounds(value.Interface().(int), min, max)
	case reflect.Float64:
		return !isOutOfBounds(value.Float(), min, max)
	case reflect.Array:
		fallthrough
	case reflect.Slice:
		return !isOutOfBounds(value.Len(), min, max)
	}

	return false
}

func isOutOfBounds[V int | float64](value V, min, max string) (result bool) {
	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.Int:
		minConverted, _ := strconv.Atoi(min)
		maxConverted, _ := strconv.Atoi(max)
		result = minConverted > any(value).(int) || maxConverted < any(value).(int)

	case reflect.Float64:
		minConverted, _ := strconv.ParseFloat(min, 64)
		maxConverted, _ := strconv.ParseFloat(max, 64)
		result = minConverted > any(value).(float64) || maxConverted < any(value).(float64)
	}
	return result
}
