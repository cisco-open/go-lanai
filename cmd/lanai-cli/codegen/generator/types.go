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
	Security Security
}

/*********************
	API Contract
 *********************/

type Contract struct {
	Path   string
	Naming ContractNaming
}

type ContractNaming struct {
	RegExps map[string]string
}

/*********************
	Web Security
 *********************/

type Security struct {
	Authentication Authentication
	Access         Access
}

type AuthenticationMethod string

const (
	AuthNone   AuthenticationMethod = `none`
	AuthOAuth2 AuthenticationMethod = `oauth2`
	// TODO more authentication methods like basic, form, etc...
)

type Authentication struct {
	Method AuthenticationMethod
}

type AccessPreset string

const (
	AccessPresetFreestyle AccessPreset = `freestyle`
	AccessPresetOPA       AccessPreset = `opa`
)

type Access struct {
	Preset AccessPreset
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
