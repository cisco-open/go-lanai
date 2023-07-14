package _go

import (
	"path"
	"strings"
)

var structRegistry = make(map[string]string)

func RegisterStruct(schemaName string, packageName string) string {
	structRegistry[strings.ToLower(schemaName)] = packageName
	return ""
}

func StructLocation(schemaName string) string {
	return structRegistry[strings.ToLower(path.Base(schemaName))]
}

func StructRegistry() map[string]string {
	return structRegistry
}
