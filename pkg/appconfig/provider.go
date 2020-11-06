package appconfig

type Provider interface {
	Load() error
	GetSettings() map[string]interface{}
	GetPrecedence() int
	isLoaded() bool
}
