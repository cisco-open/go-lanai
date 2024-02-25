// Copyright 2023 Cisco Systems, Inc. and its affiliates
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package sp

import (
    "context"
    "fmt"
    "github.com/cisco-open/go-lanai/pkg/security"
    "github.com/cisco-open/go-lanai/pkg/security/access"
    "github.com/cisco-open/go-lanai/pkg/security/errorhandling"
    "github.com/cisco-open/go-lanai/pkg/security/idp"
    "github.com/cisco-open/go-lanai/pkg/security/request_cache"
    lanaisaml "github.com/cisco-open/go-lanai/pkg/security/saml"
    "github.com/cisco-open/go-lanai/pkg/security/saml/sp/testdata"
    "github.com/cisco-open/go-lanai/pkg/web/matcher"
    "github.com/cisco-open/go-lanai/test"
    "github.com/cisco-open/go-lanai/test/apptest"
    "github.com/cisco-open/go-lanai/test/samltest"
    "github.com/cisco-open/go-lanai/test/sectest"
    "github.com/cisco-open/go-lanai/test/webtest"
    "github.com/crewjam/saml/samlsp"
    "github.com/onsi/gomega"
    "go.uber.org/fx"
    "io"
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
	idpManager := samltest.NewMockedIdpManager(func(opt *samltest.IdpManagerMockOption) {
		opt.IDPList = testdata.DefaultIdpProviders
	})
	cfg := security.ConfigurerFunc(func(ws security.WebSecurity) {
		condition := idp.RequestWithAuthenticationFlow(idp.ExternalIdpSAML, idpManager)
		ws = ws.AndCondition(condition)
		ws.Route(matcher.AnyRoute()).
			AndCondition(condition).
			With(access.New().Request(matcher.AnyRequest()).Authenticated()).
			With(New().Issuer(samltest.DefaultIssuer))
	})
	registrar.Register(&cfg)
	return MetadataTestOut{
		SecConfigurer: &cfg,
		IdpManager:    idpManager,
		AccountStore:  sectest.NewMockedFederatedAccountStore(),
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
		apptest.WithModules(Module, request_cache.Module, lanaisaml.Module, access.Module, errorhandling.Module),
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
		ctx = sectest.ContextWithSecurity(ctx)
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
	bytes, e := io.ReadAll(body)
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
