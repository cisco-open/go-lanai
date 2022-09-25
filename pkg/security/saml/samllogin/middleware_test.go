package samllogin

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/access"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/errorhandling"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/idp"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/request_cache"
	lanaisaml "cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/saml"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/saml/samllogin/testdata"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/matcher"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/apptest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/sectest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/webtest"
	"fmt"
	"github.com/crewjam/saml/samlsp"
	"github.com/onsi/gomega"
	"go.uber.org/fx"
	"io"
	"io/ioutil"
	"net/http"
	"reflect"
	"sort"
	"testing"
)

/*************************
	Setup
 *************************/

type MetadataTestOut struct {
	fx.Out
	SecConfigurer security.Configurer
	IdpManager    idp.IdentityProviderManager
	AccountStore  security.FederatedAccountStore
}

func MetadataTestSecurityConfigProvider(registrar security.Registrar) MetadataTestOut {
	idpManager := testdata.NewTestIdpManager()
	cfg := security.ConfigurerFunc(func(ws security.WebSecurity) {
		condition := idp.RequestWithAuthenticationFlow(idp.ExternalIdpSAML, idpManager)
		ws = ws.AndCondition(condition)
		ws.Route(matcher.AnyRoute()).
			AndCondition(condition).
			With(access.New().Request(matcher.AnyRequest()).Authenticated()).
			With(New().Issuer(testdata.TestIssuer))
	})
	registrar.Register(&cfg)
	return MetadataTestOut{
		SecConfigurer: &cfg,
		IdpManager:    idpManager,
		AccountStore:  testdata.NewTestFedAccountStore(),
	}
}

/*************************
	Tests
 *************************/

type metadataTestDI struct {
	fx.In
	Properties lanaisaml.SamlProperties
}

func TestMetadataEndpoint(t *testing.T) {
	di := &metadataTestDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		webtest.WithMockedServer(),
		sectest.WithMockedMiddleware(sectest.MWEnableSession()),
		apptest.WithModules(SamlAuthModule, request_cache.Module, lanaisaml.Module, access.AccessControlModule, errorhandling.ErrorHandlingModule),
		apptest.WithDI(di),
		apptest.WithConfigFS(testdata.TestConfigFS),
		apptest.WithFxOptions(
			fx.Provide(MetadataTestSecurityConfigProvider),
		),
		test.GomegaSubTest(SubTestMetadata(di), "TestMetadata"),
	)
}

/*************************
	Sub Tests
 *************************/

func SubTestMetadata(_ *metadataTestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var req *http.Request
		var resp *http.Response
		ctx = sectest.WithMockedSecurity(ctx)
		req = webtest.NewRequest(ctx, http.MethodGet, "/saml/metadata", nil)
		resp = webtest.MustExec(ctx, req).Response
		g.Expect(resp).To(MetadataMatcher{}, "metadata should return correct response")
	}
}

/*************************
	Helpers
 *************************/

type MetadataMatcher struct{}

func (m MetadataMatcher) Match(actual interface{}) (success bool, err error) {
	const (
		expectedEntityID = "http://vms.com:8080/test/saml/metadata"
		expectedACS      = "http://saml.vms.com:8080/test/saml/SSO"
		expectedSLO      = "http://saml.vms.com:8080/test/saml/slo"
	)
	body := actual.(*http.Response).Body
	bytes, e := ioutil.ReadAll(body)
	if e != nil {
		return false, e
	}
	descriptor, err := samlsp.ParseMetadata(bytes)
	if err != nil {
		return false, err
	}

	if e := m.compare(expectedEntityID, descriptor.EntityID, "entity ID"); e != nil {
		return false, nil
	}

	if e := m.compare(1, len(descriptor.SPSSODescriptors), "SP descriptors count"); e != nil {
		return false, e
	}

	if e := m.compare(2, len(descriptor.SPSSODescriptors[0].AssertionConsumerServices), "ACS bindings count"); e != nil {
		return false, e
	}

	sort.SliceStable(descriptor.SPSSODescriptors[0].AssertionConsumerServices, func(i, j int) bool {
		return descriptor.SPSSODescriptors[0].AssertionConsumerServices[i].Location > descriptor.SPSSODescriptors[0].AssertionConsumerServices[j].Location
	})
	if e := m.compare(expectedACS, descriptor.SPSSODescriptors[0].AssertionConsumerServices[0].Location, "ACS location"); e != nil {
		return false, e
	}

	if e := m.compare(2, len(descriptor.SPSSODescriptors[0].SingleLogoutServices), "SLO bindings count"); e != nil {
		return false, e
	}

	sort.SliceStable(descriptor.SPSSODescriptors[0].SingleLogoutServices, func(i, j int) bool {
		return descriptor.SPSSODescriptors[0].SingleLogoutServices[i].Location > descriptor.SPSSODescriptors[0].SingleLogoutServices[j].Location
	})
	if e := m.compare(expectedSLO, descriptor.SPSSODescriptors[0].SingleLogoutServices[0].Location, "SLO location"); e != nil {
		return false, e
	}

	return true, nil
}

func (m MetadataMatcher) compare(expected, actual interface{}, name string) error {
	if !reflect.DeepEqual(expected, actual) {
		return fmt.Errorf(`incorrect %s: expected %v, but got %v`, name, expected, actual)
	}
	return nil
}

func (m MetadataMatcher) FailureMessage(actual interface{}) (message string) {
	body := actual.(*http.Response).Body
	bytes, _ := io.ReadAll(body)
	return fmt.Sprintf("metadata doesn't match expectation. actual meta is %s", string(bytes))
}

func (m MetadataMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	body := actual.(*http.Response).Body
	bytes, _ := io.ReadAll(body)
	return fmt.Sprintf("metadata doesn't match expectation. actual meta is %s", string(bytes))
}
