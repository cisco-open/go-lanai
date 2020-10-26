package config

type Provider interface {
	GetDescription() string
	Load()
	GetSettings() map[string]interface{}
	GetPrecedence() int
	isValid() bool
}
