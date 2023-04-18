package actuator_tests

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/actuator"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/sectest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/webtest"
	"fmt"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
	"github.com/spyzhov/ajson"
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

/****************************
	Common Gomega Matchers
 ****************************/

type GomegaJsonPathMatcher struct {
	parsedJsonPath []string
	jsonPath       string
	delegate       types.GomegaMatcher
}

func MatchJsonPath(jsonPath string, matcher types.GomegaMatcher) types.GomegaMatcher {
	parsed, _ := ajson.ParseJSONPath(jsonPath)
	return &GomegaJsonPathMatcher{
		parsedJsonPath: parsed,
		jsonPath:       jsonPath,
		delegate:       matcher,
	}
}

func (m *GomegaJsonPathMatcher) Match(actual interface{}) (success bool, err error) {
	data := []byte(m.asString(actual))
	root, e := ajson.Unmarshal(data)
	if e != nil {
		return false, fmt.Errorf(`expect json string but got %T`, actual)
	}
	nodes, e := ajson.ApplyJSONPath(root, m.parsedJsonPath)
	if e != nil {
		return false, fmt.Errorf(`invalid JsonPath "%s"`, m.jsonPath)
	}

	if len(nodes) == 0 {
		return false, nil
	}
	for _, node := range nodes {
		v, e := node.Unpack()
		if e != nil {
			return false, fmt.Errorf(`unable to extract value of JsonPath [%s]: %v'`, m.jsonPath, e)
		}
		return m.delegate.Match(v)
	}
	return
}

func (m *GomegaJsonPathMatcher) FailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("JsonPath %s is\n %s", m.jsonPath, m.delegate.FailureMessage(m.asString(actual)))
}

func (m *GomegaJsonPathMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("JsonPath %s is\n %s", m.jsonPath, m.delegate.FailureMessage(m.asString(actual)))
}

func (m *GomegaJsonPathMatcher) asString(actual interface{}) string {
	var data string
	switch v := actual.(type) {
	case string:
		data = v
	case []byte:
		data = string(v)
	}
	return data
}
