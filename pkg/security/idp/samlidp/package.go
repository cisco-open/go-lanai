package samlidp

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/saml/samllogin"
)

//var logger = log.New("SEC.SAML")

func Use() {
	samllogin.Use() // samllogin enables External SAML IDP support
}

