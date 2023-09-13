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

func (c Config) ToV2() *ConfigV2 {
	regenRules := make([]RegenRule, 0, len(c.Regeneration.Rules))
	for k, v := range c.Regeneration.Rules {
		regenRules = append(regenRules, RegenRule{
			Pattern: k,
			Mode:    RegenMode(v),
		})
	}
	return &ConfigV2{
		Project:    ProjectV2{
			Name:   c.ProjectName,
			Module: c.RepositoryRootPath,
		},
		Templates:  TemplatesV2{
			Path: c.TemplateDirectory,
		},
		Components: ComponentsV2{
			Contract: ContractV2{
				Path: c.Contract,
				Naming: ContractNamingV2{
					RegExps: c.Regexes,
				},
			},
		},
		Regen:      RegenerationV2{
			Default: RegenMode(c.Regeneration.Default),
			Rules:   regenRules,
		},
	}
}