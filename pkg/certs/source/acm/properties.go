package acmcerts

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
)

type SourceProperties struct {
	MinTLSVersion    string         `json:"min-version"`
	ARN              string         `json:"arn"`
	Passphrase       string         `json:"passphrase"`
	MinRenewInterval utils.Duration `json:"min-renew-interval"`
	CachePath        string         `json:"cache-path"`
}