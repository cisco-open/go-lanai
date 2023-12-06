package tlsconfig

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"encoding/json"
	"github.com/pkg/errors"
)

const PropertiesPrefix = `tls`

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

type FileCache struct {
	Enabled bool   `json:"enabled"`
	Path    string `json:"path"`
	Prefix  string `json:"prefix"`
}

// BindProperties create and bind SessionProperties, with a optional prefix
func BindProperties(appCfg bootstrap.ApplicationConfig) Properties {
	var props Properties
	if err := appCfg.Bind(&props, PropertiesPrefix); err != nil {
		panic(errors.Wrap(err, "failed to bind Properties"))
	}
	return props
}