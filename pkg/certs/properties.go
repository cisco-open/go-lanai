// Copyright 2023 Cisco Systems, Inc. and its affiliates
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

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
