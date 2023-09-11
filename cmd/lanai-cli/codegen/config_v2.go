package codegen

import "cto-github.cisco.com/NFV-BU/go-lanai/cmd/lanai-cli/codegen/generator"

type ConfigV2 struct {
	Project    ProjectV2      `yaml:"project"`
	Templates  TemplatesV2    `yaml:"templates"`
	Components ComponentsV2   `yaml:"components"`
	Regen      RegenerationV2 `yaml:"regen"`
}

type ProjectV2 struct {
	// Name service name
	Name string `yaml:"name"`
	// Module golang module
	Module string `yaml:"module"`
}

type TemplatesV2 struct {
	Path string `yaml:"path"`
}

type ComponentsV2 struct {
	Contract ContractV2 `yaml:"contract"`
}

type ContractV2 struct {
	Path   string           `yaml:"path"`
	Naming ContractNamingV2 `yaml:"naming"`
}

type ContractNamingV2 struct {
	RegExps map[string]string `yaml:"regular-expressions"`
}

type RegenMode generator.RegenMode

type RegenRule struct {
	// Pattern wildcard pattern of output file path
	Pattern string `yaml:"pattern"`
	// Mode regeneration mode on matched output files in case of changes. (ignore, overwrite, reference, etc.)
	Mode RegenMode `yaml:"mode"`
}

type RegenRules []RegenRule

type RegenerationV2 struct {
	Default RegenMode  `yaml:"default"`
	Rules   RegenRules `yaml:"rules"`
}

func (r RegenerationV2) AsGeneratorOption() func(*generator.Option) {
	rules := make(generator.RegenRules, len(r.Rules))
	for i := range r.Rules {
		rules[i] = generator.RegenRule{
			Pattern: r.Rules[i].Pattern,
			Mode:    generator.RegenMode(r.Rules[i].Mode),
		}
	}
	return generator.WithRegenRules(rules, generator.RegenMode(r.Default))
}
