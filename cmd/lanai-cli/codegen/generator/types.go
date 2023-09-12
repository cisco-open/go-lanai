package generator

/******************
	Context Data
 ******************/

// Keys in template's context data as map
const (
	CKOpenAPIData = "OpenAPIData"
	CKProjectName = "ProjectName"
	CKRepository  = "Repository"
	CKProject     = "Project"
)

type Project struct {
	Name        string
	Module      string
	Description string
	Port        int
	ContextPath string
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
