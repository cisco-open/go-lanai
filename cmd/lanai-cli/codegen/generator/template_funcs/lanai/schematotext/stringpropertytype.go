package schematotext

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/cmd/lanai-cli/codegen/generator/template_funcs/lanai/lanaiutil"
	"github.com/google/uuid"
	"reflect"
	"time"
)

type stringType struct {
	data                 interface{}
	baseProcessor        *baseType
	externalTypesAllowed bool
}

func NewStringType(data interface{}, opts ...func(option *translatorOptions)) *stringType {
	o := &translatorOptions{}
	for _, fn := range opts {
		fn(o)
	}

	return &stringType{
		data:                 data,
		baseProcessor:        NewBaseType(data, opts...),
		externalTypesAllowed: !o.restrictExternalTypes,
	}
}
func (s stringType) ToText() (string, error) {
	result, err := s.baseProcessor.ToText()
	if err != nil {
		return "", err
	}
	if s.externalTypesAllowed {
		//Perform modifications on the base type depending on the format
		if lanaiutil.MatchesFormat(s.data, "uuid") {
			result = reflect.TypeOf(uuid.UUID{}).String()
		} else if lanaiutil.MatchesFormat(s.data, "date-time") {
			result = reflect.TypeOf(time.Time{}).String()
		}
	}
	return result, nil
}
