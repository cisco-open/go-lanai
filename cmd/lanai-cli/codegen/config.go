package codegen

type ConfigVersion string

const (
	VersionUnknown ConfigVersion = ``
	Version1       ConfigVersion = `v1`
	Version2       ConfigVersion = `v2`
)

type VersionedConfig struct {
	Version ConfigVersion `yaml:"version"`
	Config
	ConfigV2
}

type Regeneration struct {
	Default string            `yaml:"default"`
	Rules   map[string]string `yaml:"rules"`
}

type Config struct {
	Contract           string            `yaml:"contract"`
	ProjectName        string            `yaml:"projectName"`
	TemplateDirectory  string            `yaml:"templateDirectory"`
	RepositoryRootPath string            `yaml:"repositoryRootPath"`
	Regeneration       Regeneration      `yaml:"regeneration"`
	Regexes            map[string]string `yaml:"regexes"`
}
