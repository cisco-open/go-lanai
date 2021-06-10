package vaultprovider

const (
	KvConfigPrefix = "cloud.vault.kv"
)

// KvConfigProperties currently only supports v1 kv secret engine
// TODO review property path and prefix
type KvConfigProperties struct {
	Enabled     bool `json:"enabled"`
	Backend          string `json:"backend"`
	DefaultContext   string `json:"default-context"`
	ProfileSeparator string `json:"profile-separator"`
	ApplicationName  string `json:"application-name"`
	BackendVersion   int    `json:"backend-version"`
}

