package generator

/*********************
	Project
 *********************/

type Project struct {
	Name        string
	Module      string
	Description string
	Port        int
	ContextPath string
}

/*********************
	Components
 *********************/

type Components struct {
	Contract Contract
}

type Contract struct {
	Path   string
	Naming ContractNaming
}

type ContractNaming struct {
	RegExps map[string]string
}

/********************
	Template Data
 ********************/

// Keys in template's context data as map
const (
	KDataOpenAPI     = "OpenAPIData"
	KDataProjectName = "ProjectName"
	KDataRepository  = "Repository"
	KDataProject     = "Project"
)

func newCommonData(p *Project) map[string]interface{} {
	return map[string]interface{}{
		KDataProjectName: p.Name,
		KDataRepository: p.Module,
		KDataProject: p,
	}
}

/******************
	Regen
 ******************/

// RegenMode file operation mode when re-generating.
type RegenMode string

const (
	RegenModeIgnore    RegenMode = "ignore"
	RegenModeReference RegenMode = "reference"
	RegenModeOverwrite RegenMode = "overwrite"
)

// RegenRule file operation rules during re-generation
type RegenRule struct {
	// Pattern wildcard pattern of output file path
	Pattern string
	// Mode regeneration mode on matched output files in case of changes. (ignore, overwrite, reference, etc.)
	Mode RegenMode
}

type RegenRules []RegenRule
