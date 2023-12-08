package certs

import (
	"encoding/json"
)

type Properties struct {
	Sources map[SourceType]json.RawMessage `json:"sources"`
	Presets map[string]json.RawMessage     `json:"presets"`
}

// SourceProperties convenient properties for other package to bind
type SourceProperties struct {
	Preset string          `json:"preset"`
	Raw    json.RawMessage `json:"-"`
}

func (p *SourceProperties) UnmarshalJSON(data []byte) error {
	p.Raw = data
	type props SourceProperties
	return json.Unmarshal(data, (*props)(p))
}

func NewProperties() *Properties {
	return &Properties{
		Sources: map[SourceType]json.RawMessage{},
		Presets: map[string]json.RawMessage{},
	}
}
