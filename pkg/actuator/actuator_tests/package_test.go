package actuator_tests

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/actuator"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/sectest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/webtest"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
	"net/http"
	"testing"
)

/*************************
	Common Test Setup
 *************************/

func ConfigureSecurity(reg *actuator.Registrar) {
	reg.MustRegister(actuator.SecurityCustomizerFunc(func(ws security.WebSecurity) {}))
}

type TestDI struct {
	fx.In
}

/*************************
	Common Helpers
 *************************/

func mockedSecurityAdmin() sectest.SecurityContextOptions {
	return sectest.MockedAuthentication(func(d *sectest.SecurityDetailsMock) {
		d.Permissions = utils.NewStringSet("IS_API_ADMIN")
	})
}

func mockedSecurityNonAdmin() sectest.SecurityContextOptions {
	return sectest.MockedAuthentication(func(d *sectest.SecurityDetailsMock) {
		d.Permissions = utils.NewStringSet("not_worthy")
	})
}

func defaultRequestOptions() webtest.RequestOptions {
	return webtest.Headers(
		"Accept", "application/json",
	)
}

func v2RequestOptions() webtest.RequestOptions {
	return webtest.Headers(
		"Accept", actuator.ContentTypeSpringBootV2,
	)
}

func assertResponse(_ *testing.T, g *gomega.WithT, resp *http.Response, expectedStatus int, expectedHeaders ...string) {
	g.Expect(resp).ToNot(BeNil(), "endpoint should have response")
	g.Expect(resp.StatusCode).To(BeEquivalentTo(expectedStatus))
	for i := range expectedHeaders {
		if i%2 == 1 || i+1 >= len(expectedHeaders) {
			continue
		}
		k := expectedHeaders[i]
		v := expectedHeaders[i+1]
		g.Expect(resp.Header.Get(k)).To(BeEquivalentTo(v), "response header should contains [%s]='%s'", k, v)
	}
}
