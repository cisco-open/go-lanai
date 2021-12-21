package actuator

import "cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"

const (
	MinActuatorPrecedence = bootstrap.ActuatorPrecedence
	MaxActuatorPrecedence = bootstrap.ActuatorPrecedence + bootstrap.FrameworkModulePrecedenceBandwidth
)

const (
	ContentTypeSpringBootV2 = "application/vnd.spring-boot.actuator.v2+json"
	ContentTypeSpringBootV3 = "application/vnd.spring-boot.actuator.v3+json"
)
