// Package certs
// Defines necessary interfaces and types for certificate management
package certs

import (
	"context"
	"crypto/tls"
	"encoding/json"
)

const (
	FxGroup = "certs"
)

const (
	SourceVault SourceType = "vault"
	SourceFile  SourceType = "file"
	SourceACM   SourceType = "acm"
)
type SourceType string

type TLSOptions func(opt *TLSOption)
type TLSOption struct {
	// For now, there is no configurable options
}

type Source interface {
	// TLSConfig get certificates as tls.Config. For native drivers that support standard tls.Config
	TLSConfig(ctx context.Context, opts ...TLSOptions) (*tls.Config, error)
	// Files get certificates as local files. For drivers that support filesystem based certificates config e.g. postgres DSN
	Files(ctx context.Context) (*CertificateFiles, error)
}

// CertificateFiles filesystem based certificates and keys.
// All values in this struct are corresponding file's path on local filesystem.
// Some system can only reference certificates by path on filesystem
type CertificateFiles struct {
	RootCAPaths          []string
	CertificatePath      string
	PrivateKeyPath       string
	PrivateKeyPassphrase string
}

type Options func(opt *Option)
type Option struct {
	// Preset name of the preset config. Set this field to reuse configuration from properties (tls.presets.<name>).
	// This field is exclusive with ConfigPath, Type and RawConfig
	Preset string

	// ConfigPath is similar to Preset, but should be the full property path. e.g.  "redis.tls.config".
	// This field is exclusive with Preset, Type and RawConfig
	ConfigPath string

	// RawConfig raw configuration of the certificate source, required when Type is set.
	// This field is exclusive with Preset and ConfigPath
	// Supported types: json.RawMessage, []byte (JSON), string (JSON), or any struct compatible with corresponding SourceType
	RawConfig interface{}

	// Type type of the certificate source. Set this field for manual configuration
	// This field is ignored if any of Preset or ConfigPath is set.
	// If RawConfig includes "type" field, Type is optional. In such case, if Type is set, it overrides the value from RawConfig
	Type SourceType
}

// Manager is the package's top-level interface that provide TLS configurations
type Manager interface {
	Source(ctx context.Context, opts ...Options) (Source, error)
}

// Registrar is the additional top-level interface for supported Provider to register itself
// Supported types:
// - SourceFactory
type Registrar interface {
	Register(items ...interface{}) error
	MustRegister(items ...interface{})
}

type SourceOptions func(srcCfg *SourceConfig)
type SourceConfig struct {
	RawConfig json.RawMessage
}

type SourceFactory interface {
	Type() SourceType
	LoadAndInit(ctx context.Context, opts ...SourceOptions) (Source, error)
}

// ProviderFactory
// Deprecated
//type ProviderFactory struct {
//	Manager
//	AppCtx *bootstrap.ApplicationContext
//}
//
//func (f *ProviderFactory) GetProvider(properties Properties) (Provider, error) {
//	opts, e := f.convertLegacyProps(&properties)
//	if e != nil {
//		return nil, e
//	}
//	if src, e := f.Manager.Source(f.AppCtx, opts); e != nil {
//		return nil, e
//	} else {
//		return src.(Provider), nil
//	}
//}
//
//func (f *ProviderFactory) convertLegacyProps(properties *Properties) (Options, error) {
//	rawConfig, e := json.Marshal(properties)
//	if e != nil {
//		return nil, fmt.Errorf(`cannot convert legacy properties: %v`, e)
//	}
//	return func(opt *Option) {
//		opt.Type = properties.Type
//		opt.RawConfig = rawConfig
//	}, nil
//}
