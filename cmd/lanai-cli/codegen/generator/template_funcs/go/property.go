package _go

// Property represents strictly what is needed to represent a struct property in Golang
type Property struct {
	Name     string
	Type     string
	Bindings string
}

func NewProperty(name string, typeOfProperty string) *Property {
	return &Property{
		Name: name,
		Type: typeOfProperty,
	}
}

func (m *Property) AddBinding(binding string) *Property {
	m.Bindings = binding
	return m
}

func (m *Property) AddType(typeOfProperty string) *Property {
	m.Type = typeOfProperty
	return m
}
