package appconfig

type Provider interface {
	Name() string
	Load() error
	GetSettings() map[string]interface{}
	GetPrecedence() int
	IsLoaded() bool
}
