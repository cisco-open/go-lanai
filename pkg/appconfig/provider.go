package appconfig

type Provider interface {
	GetDescription() string
	Load() error
	GetSettings() map[string]interface{}
	GetPrecedence() int
	isLoaded() bool
}
