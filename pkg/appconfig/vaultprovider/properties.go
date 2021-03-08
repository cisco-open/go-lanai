package vaultprovider

const (
	KvConfigPrefix = "cloud.vault.generic"
)

type KvConfigProperties struct {
	Enabled     string `json:"enabled"`
	Backend          string `json:"secretEngine"`
	DefaultContext   string `json:"default-context"`
	ProfileSeparator string `json:"profile-separator"`
	ApplicationName  string `json:"application-name"`
}

