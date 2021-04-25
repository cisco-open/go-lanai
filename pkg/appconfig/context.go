package appconfig

import "cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"

const (
	PropertyKeyActiveProfiles       = "application.profiles.active"
	PropertyKeyAdditionalProfiles   = "application.profiles.additional"
	PropertyKeyConfigFileSearchPath = "config.file.search-path"
	PropertyKeyApplicationName      = bootstrap.PropertyKeyApplicationName
	PropertyKeyBuildInfo            = "application.build"
	//PropertyKey = ""
)

type ConfigAccessor interface {
	bootstrap.ApplicationConfig
	Each(apply func(string, interface{}) error) error
	// Providers gives effective config providers
	Providers() []Provider
	Profiles() []string
	HasProfile(profile string) bool
}

type BootstrapConfig struct {
	config
}

func NewBootstrapConfig(groups ...ProviderGroup) *BootstrapConfig {
	return &BootstrapConfig{config: config{groups: groups}}
}

type ApplicationConfig struct {
	config
}

func NewApplicationConfig(groups ...ProviderGroup) *ApplicationConfig {
	return &ApplicationConfig{config: config{groups: groups}}
}

