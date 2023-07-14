package opa

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
)

const PropertiesPrefix = "security.opa"

type Properties struct {
	Server  BundleServerProperties            `json:"server"`
	Bundles map[string]BundleSourceProperties `json:"bundles"`
	Logging LoggingProperties                 `json:"logging"`
}

type BundleServerProperties struct {
	Name string `json:"name"`
	URL  string `json:"url"`
	PollingProperties
}

type BundleSourceProperties struct {
	Path               string         `json:"path"`
	PollingProperties
}

type LoggingProperties struct {
	DecisionLogsEnabled bool `json:"decision-logs-enabled"`
}

type PollingProperties struct {
	PollingMinDelay    *utils.Duration `json:"polling-min-delay,omitempty"`    // min amount of time to wait between successful poll attempts
	PollingMaxDelay    *utils.Duration `json:"polling-max-delay,omitempty"`    // max amount of time to wait between poll attempts
	LongPollingTimeout *utils.Duration `json:"long-polling-timeout,omitempty"` // max amount of time the server should wait before issuing a timeout if there's no update available
}

func NewProperties() *Properties {
	return &Properties{}
}
