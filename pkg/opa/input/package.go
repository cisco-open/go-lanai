package opainput

import "cto-github.cisco.com/NFV-BU/go-lanai/pkg/opa"

var DefaultInputCustomizers = []opa.InputCustomizer{
	opa.InputCustomizerFunc(PopulateAuthenticationClause),
}
