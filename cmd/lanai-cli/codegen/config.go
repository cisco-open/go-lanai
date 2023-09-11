package codegen

type ConfigVersion string

const (
	VersionUnknown ConfigVersion = ``
	Version1       ConfigVersion = `v1`
	Version2       ConfigVersion = `v2`
)

type VersionedConfig struct {
	Version ConfigVersion `json:"version"`
	Config
	ConfigV2
}

type Regeneration struct {
	Default string            `json:"default"`
	Rules   map[string]string `json:"rules"`
}

type Config struct {
	Contract           string            `json:"contract"`
	ProjectName        string            `json:"projectName"`
	TemplateDirectory  string            `json:"templateDirectory"`
	RepositoryRootPath string            `json:"repositoryRootPath"`
	Regeneration       Regeneration      `json:"regeneration"`
	Regexes            map[string]string `json:"regexes"`
}
