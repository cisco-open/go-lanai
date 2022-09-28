package testdata

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/webtest"
)

var TestIssuer = security.NewIssuer(func(opt *security.DefaultIssuerDetails) {
	*opt =security.DefaultIssuerDetails{
		Protocol:    "http",
		Domain:      "vms.com",
		Port:        8080,
		ContextPath: webtest.DefaultContextPath,
		IncludePort: true,
	}})
