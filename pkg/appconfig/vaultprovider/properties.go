package vaultprovider

const (
	KvConfigPrefix = "cloud.vault.kv"
)

//currently only supports v1 kv secret engine
type KvConfigProperties struct {
	Enabled     bool `json:"enabled"`
	Backend          string `json:"backend"`
	DefaultContext   string `json:"default-context"`
	ProfileSeparator string `json:"profile-separator"`
	ApplicationName  string `json:"application-name"`
	BackendVersion   int    `json:"backend-version"`
}

