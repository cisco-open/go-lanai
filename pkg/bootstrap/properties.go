package bootstrap

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
)

type Properties struct {
	Application ApplicationProperties `json:"application"`
	Cloud       CloudProperties       `json:"cloud"`
}

type ApplicationProperties struct {
	Name     string            `json:"name"`
	Profiles ProfileProperties `json:"profiles"`
}

type ProfileProperties struct {
	Active     utils.CommaSeparatedSlice `json:"active"`
	Additional utils.CommaSeparatedSlice `json:"additional"`
}

type CloudProperties struct {
	Gateway GatewayProperties `json:"gateway"`
}

type GatewayProperties struct {
	Service string `json:"service"`
	Scheme  string `json:"scheme"`
	Host    string `json:"host"`
	Port    int    `json:"port"`
}

