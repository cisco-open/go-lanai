package _go

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/cmd/lanai-cli/codegen/generator/template_funcs/util"
)

type Struct struct {
	Properties      []Property
	EmbeddedStructs []Struct
	Package         string
	Name            string
}

func NewStruct(name string, pkg string) *Struct {
	if name == "" {
		return nil
	}
	return &Struct{
		Name:    name,
		Package: pkg,
	}
}

func (m *Struct) AddProperties(property Property) *Struct {
	if m == nil {
		return nil
	}
	m.Properties = append(m.Properties, property)
	return m
}

func (m *Struct) AddEmbeddedStruct(ref string) *Struct {
	if m == nil {
		return nil
	}

	name := util.ToTitle(util.BasePath(ref))
	m.EmbeddedStructs = append(m.EmbeddedStructs, Struct{
		Package: StructLocation(name),
		Name:    name,
	})

	return m
}
