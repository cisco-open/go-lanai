package internal

import (
	"fmt"
	"path"
	"strings"
)

type DefaultType struct {
	data              interface{}
	defaultObjectName string
	currentPkg        string
}

func NewDefaultType(data interface{}, opts ...func(option *TranslatorOptions)) *DefaultType {
	o := &TranslatorOptions{}
	for _, fn := range opts {
		fn(o)
	}

	return &DefaultType{
		data:              data,
		defaultObjectName: o.defaultObjectName,
		currentPkg:        o.currentPkg,
	}
}
func (s DefaultType) toText() (string, error) {
	schema, err := convertToSchemaRef(s.data)
	if err != nil {
		return "", err
	}
	result := ""
	if schema == nil || schema.Ref == "" {
		result = s.defaultObjectName
	} else {
		result = path.Base(schema.Ref)
		refPackage, ok := structRegistry[strings.ToLower(result)]
		if ok && refPackage != s.currentPkg {
			result = fmt.Sprintf("%v.%v", path.Base(refPackage), result)
		}
	}
	return result, nil
}
