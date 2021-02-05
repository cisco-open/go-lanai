package example

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/idp"
	"errors"
)

type InMemAuthFlowManager struct {

}

func (i *InMemAuthFlowManager) GetAuthFlow(domain string) (string, error) {
	if domain == "saml.vms.com" {
		return idp.ExternalIdpSAML, nil
	}
	return "", errors.New("no auth flow defined")
}

func NewInMemAuthFlowManager() idp.AuthFlowManager{
	return &InMemAuthFlowManager{}
}

