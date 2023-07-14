package schematotext

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/cmd/lanai-cli/codegen/generator/template_funcs/go"
	"cto-github.cisco.com/NFV-BU/go-lanai/cmd/lanai-cli/codegen/generator/template_funcs/lanai/lanaiutil"
	"fmt"
	"path"
	"strings"
)

type defaultType struct {
	data              interface{}
	defaultObjectName string
	currentPkg        string
}

func NewDefaultType(data interface{}, opts ...func(option *translatorOptions)) *defaultType {
	o := &translatorOptions{}
	for _, fn := range opts {
		fn(o)
	}

	return &defaultType{
		data:              data,
		defaultObjectName: o.defaultObjectName,
		currentPkg:        o.currentPkg,
	}
}
func (s defaultType) ToText() (string, error) {
	schema, err := lanaiutil.ConvertToSchemaRef(s.data)
	if err != nil {
		return "", err
	}
	result := ""
	if schema == nil || schema.Ref == "" {
		result = s.defaultObjectName
	} else {
		result = path.Base(schema.Ref)
		refPackage, ok := _go.StructRegistry()[strings.ToLower(result)]
		if ok && refPackage != s.currentPkg {
			result = fmt.Sprintf("%v.%v", path.Base(refPackage), result)
		}
	}
	return result, nil
}
