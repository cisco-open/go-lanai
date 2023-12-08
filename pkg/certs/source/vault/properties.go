package vaultcerts

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
)

type SourceProperties struct {
	MinTLSVersion    string         `json:"min-version"`
	Path             string         `json:"path"`
	Role             string         `json:"role"`
	CN               string         `json:"cn"`
	IpSans           string         `json:"ip-sans"`
	AltNames         string         `json:"alt-names"`
	TTL              string         `json:"ttl"`
	MinRenewInterval utils.Duration `json:"min-renew-interval"`
	CachePath        string         `json:"cache-path"`
}
