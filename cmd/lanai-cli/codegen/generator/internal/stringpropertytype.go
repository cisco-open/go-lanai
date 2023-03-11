package internal

import (
	"github.com/google/uuid"
	"reflect"
	"time"
)

type StringType struct {
	data                 interface{}
	baseProcessor        *BaseType
	externalTypesAllowed bool
}

func NewStringType(data interface{}, opts ...func(option *TranslatorOptions)) *StringType {
	o := &TranslatorOptions{}
	for _, fn := range opts {
		fn(o)
	}

	return &StringType{
		data:                 data,
		baseProcessor:        NewBaseType(data, opts...),
		externalTypesAllowed: !o.restrictExternalTypes,
	}
}
func (s StringType) toText() (string, error) {
	result, err := s.baseProcessor.toText()
	if err != nil {
		return "", err
	}
	if s.externalTypesAllowed {
		//Perform modifications on the base type depending on the format
		if matchesFormat(s.data, "uuid") {
			result = reflect.TypeOf(uuid.UUID{}).String()
		} else if matchesFormat(s.data, "date-time") {
			result = reflect.TypeOf(time.Time{}).String()
		}
	}
	return result, nil
}
