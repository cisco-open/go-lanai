package actuator

import "cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"

const (
	MinActuatorPrecedence = bootstrap.FrameworkModulePrecedence + 3000
	MaxActuatorPrecedence = bootstrap.FrameworkModulePrecedence + 3999
)
