package tlsconfig

import (
	"encoding/json"
)

type Properties struct {
	// type can be vault or file
	Type       SourceType `json:"type"`
	MinVersion string     `json:"min-version"`

	// vault type related properties
	Path             string    `json:"path"`
	Role             string    `json:"role"`
	CN               string    `json:"cn"`
	IpSans           string    `json:"ip-sans"`
	AltNames         string    `json:"alt-names"`
	TTL              string    `json:"ttl"`
	MinRenewInterval string    `json:"min-renew-interval"`
	FileCache        FileCache `json:"file-cache"`

	// file type related properties
	CACertFile string `json:"ca-cert-file"`
	CertFile   string `json:"cert-file"`
	KeyFile    string `json:"key-file"`
	KeyPass    string `json:"key-pass"`

	// acm type related properties
	ARN        string `json:"arn"`
	Passphrase string `json:"passphrase"`

	Sources map[SourceType]json.RawMessage `json:"sources"`
	Presets map[string]json.RawMessage     `json:"presets"`
}

// SourceProperties convenient properties for other package to bind
type SourceProperties struct {
	Preset string `json:"preset"`
	Raw json.RawMessage `json:"-"`
}

func (p *SourceProperties) UnmarshalJSON(data []byte) error {
	p.Raw = data
	type props SourceProperties
	return json.Unmarshal(data, (*props)(p))
}

type FileCache struct {
	Enabled bool   `json:"enabled"`
	Path    string `json:"path"`
	Prefix  string `json:"prefix"`
}



