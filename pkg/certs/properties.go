package certs

import (
	"encoding/json"
)

type Properties struct {
	Sources map[SourceType]json.RawMessage `json:"sources"`
	Presets map[string]json.RawMessage     `json:"presets"`
}

// SourceProperties convenient properties for other package to bind.
type SourceProperties struct {
	// Preset is optional. When set, it should match a key in Properties.Presets
	Preset string `json:"preset"`
	// Type is required when Preset is not set, optional and ignored when Preset is set.
	Type SourceType `json:"type"`
	// Raw stores configuration as JSON.
	// When Preset is set, Raw might be empty. Otherwise, Raw should at least have "type"
	Raw json.RawMessage `json:"-"`
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
