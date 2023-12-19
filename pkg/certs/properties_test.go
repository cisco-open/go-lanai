package certs_test

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/certs"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
)

// No test cases here, just bunch of test properties. See Manager tests

type TestRedisProperties struct {
	Address string            `json:"addrs"`
	TLS     TestTLSProperties `json:"tls"`
}

type TestDataProperties struct {
	Host string            `json:"host"`
	Port int               `json:"port"`
	TLS  TestTLSProperties `json:"tls"`
}

type TestKafkaProperties struct {
	Brokers utils.CommaSeparatedSlice `json:"brokers"`
	TLS     TestTLSProperties         `json:"tls"`
}

type TestTLSProperties struct {
	Enabled bool                   `json:"enabled"`
	Certs   certs.SourceProperties `json:"certs"`
}

type TestVaultSourceProperties struct {
	MinTLSVersion    string         `json:"min-version"`
	Path             string         `json:"path"`
	Role             string         `json:"role"`
	CN               string         `json:"cn"`
	IpSans           string         `json:"ip-sans"`
	AltNames         string         `json:"alt-names"`
	TTL              utils.Duration `json:"ttl"`
	MinRenewInterval utils.Duration `json:"min-renew-interval"`
	CachePath        string         `json:"cache-path"`
}

type TestFileSourceProperties struct {
	MinTLSVersion string `json:"min-version"`
	CACertFile    string `json:"ca-cert-file"`
	CertFile      string `json:"cert-file"`
	KeyFile       string `json:"key-file"`
	KeyPass       string `json:"key-pass"`
}
