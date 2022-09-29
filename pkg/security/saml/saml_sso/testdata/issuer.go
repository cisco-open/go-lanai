package testdata

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/cryptoutils"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/webtest"
)

const TestIDPCertFile   = `testdata/saml_test.cert`
var TestIDPCerts, _ = cryptoutils.LoadCert(TestIDPCertFile)

var TestIssuer = security.NewIssuer(func(opt *security.DefaultIssuerDetails) {
	*opt =security.DefaultIssuerDetails{
		Protocol:    "http",
		Domain:      "vms.com",
		Port:        8080,
		ContextPath: webtest.DefaultContextPath,
		IncludePort: true,
	}})
