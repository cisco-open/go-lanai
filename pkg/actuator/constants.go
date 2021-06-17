package actuator

import "cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"

const (
	MinActuatorPrecedence = bootstrap.ActuatorPrecedence
	MaxActuatorPrecedence = bootstrap.ActuatorPrecedence + bootstrap.FrameworkModulePrecedenceBandwidth
)
