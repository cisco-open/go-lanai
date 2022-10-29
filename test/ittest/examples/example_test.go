package examples

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/integrate/httpclient"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/integrate/security/seclient"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/apptest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/ittest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/sdtest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/sectest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/webtest"
	"encoding/json"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
	"net/http"
	"strings"
	"testing"
)

/*************************
	Tests
 *************************/

// TestMain globally enable HTTP recording mode
// Important: this should always be disabled when checking in
//func TestMain(m *testing.M) {
//	suitetest.RunTests(m,
//		ittest.PackageHttpRecordingMode(),
//	)
//}

type TestDI struct {
	fx.In
	Service    *ExampleService
	AuthClient seclient.AuthenticationClient
}


// TestExampleMockedServerTestWithSecurity
func TestExampleMockedServerTestWithSecurity(t *testing.T) {
	var di TestDI
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		// Setup your test, includes modules you need, provide test subjects and other mocks
		webtest.WithMockedServer(),
		apptest.WithModules(httpclient.Module),
		apptest.WithDI(&di),
		apptest.WithFxOptions(
			web.FxControllerProviders(NewExampleController),
			fx.Provide(NewExampleService),
		),

		// Tell test framework to use recorded HTTP interaction.
		// Note: this function accept may options. See ittest/httpvcr.go for more details
		ittest.WithHttpPlayback(t),

		// Tell test framework to use real service for any HTTP interaction.
		// This should be enabled during development and turned off before checking in the code
		//ittest.WithHttpPlayback(t, ittest.HttpRecordingMode()),

		// Because the test subjects (ExampleController, ExampleService) uses service discovery and scopes,
		// They need to be configured properly for HTTP recorder to work
		// Note: We use ittest.WithRecordedScopes instead of sectest.WithMockedScopes, because switching context
		// 		 in recording mode requires real access token in each scope. The HTTP interactions during context
		// 		 switching is also recorded
		ittest.WithRecordedScopes(),
		// See `sdtest` for more examples. The SD mocking is defined in testdata/application-test.yml
		sdtest.WithMockedSD(sdtest.DefinitionWithPrefix("mocks.sd")),

		// Because remote access require current context to be authentic during recording mode,
		// We need this configuration to pass along the security context from the test context to Controller's context.
		sectest.WithMockedMiddleware(),

		// Controller requires permissions, we need to mock it for any test that uses webtest
		test.SubTestSetup(MockPermissions()),

		// Test order is important, unless ittest.DisableHttpRecordOrdering option is used
		test.GomegaSubTest(SubTestMockedServerWithSystemAccount(), "TestMockedServerWithSystemAccount"),
		test.GomegaSubTest(SubTestMockedServerWithoutSystemAccount(&di), "TestMockedServerWithoutSystemAccount"),
		test.GomegaSubTest(SubTestMockedServerWithCurrentContext(&di), "TestMockedServerWithCurrentContext"),
	)
}

func TestExampleUnitTestWithSecurity(t *testing.T) {
	var di TestDI
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		// Setup your test, includes modules you need, provide test subjects and other mocks
		apptest.WithModules(httpclient.Module),
		apptest.WithDI(&di),
		apptest.WithFxOptions(
			fx.Provide(NewExampleService),
		),

		// Tell test framework to use recorded HTTP interaction.
		// Note: this function accept may options. See ittest/httpvcr.go for more details
		ittest.WithHttpPlayback(t),

		// Tell test framework to use real service for any HTTP interaction.
		// This should be enabled during development and turned off before checking in the code
		//ittest.WithHttpPlayback(t, ittest.HttpRecordingMode()),

		// Because the test subjects (ExampleService) uses service discovery and scopes,
		// They need to be configured properly for HTTP recorder to work
		// Note: We use ittest.WithRecordedScopes instead of sectest.WithMockedScopes, because switching context
		// 		 in recording mode requires real access token in each scope. The HTTP interactions during context
		// 		 switching is also recorded
		ittest.WithRecordedScopes(),
		// See `sdtest` for more examples. The SD mocking is defined in testdata/application-test.yml
		sdtest.WithMockedSD(sdtest.DefinitionWithPrefix("mocks.sd")),

		// Test order is important, unless ittest.DisableHttpRecordOrdering option is used
		test.GomegaSubTest(SubTestUnitTestWithSystemAccount(&di), "TestUnitTestWithSystemAccount"),
		test.GomegaSubTest(SubTestUnitTestWithWithoutSystemAccount(&di), "TestUnitTestWithWithoutSystemAccount"),
		test.GomegaSubTest(SubTestUnitTestWithCurrentContext(&di), "TestUnitTestWithCurrentContext"),
	)
}

/*************************
	Sub Tests
 *************************/

// A permission setup
func MockPermissions() test.SetupFunc {
	return func(ctx context.Context, t *testing.T) (context.Context, error) {
		// Authentication is passed along to the controller in MockedServer mode.
		return sectest.ContextWithSecurity(ctx, sectest.MockedAuthentication(func(d *sectest.SecurityDetailsMock) {
			d.Permissions = utils.NewStringSet("DUMMY_PERMISSION")
		})), nil
	}
}

func SubTestMockedServerWithSystemAccount() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		req := webtest.NewRequest(ctx, http.MethodGet, "/remote", nil,
			webtest.Queries("sys_acct", "true"),
			webtest.Queries("user", "admin"),
		)
		resp := webtest.MustExec(ctx, req).Response
		assertHttpResponse(t, g, resp, "admin")
	}
}

func SubTestUnitTestWithSystemAccount(di *TestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		ret, e := di.Service.CallRemoteWithSystemAccount(ctx, "admin")
		g.Expect(e).To(Succeed(), "functions that calling remote service should not fail")
		assertResult(t, g, ret, "admin")
	}
}

// SubTestMockedServerWithoutSystemAccount
// Because the test subject (the controller) uses current context,
// a real access token should be used in HTTP recording mode. (It doesn't matter in replay/playback mode).
// Therefore, we need additional setup to inject real access token into context.
// How real access token is obtained is irrelevant:
// 1. seclient.AuthenticationClient can be used if this user have password
// 2. A hard corded token also works, as long as it's valid at the time of recording
func SubTestMockedServerWithoutSystemAccount(di *TestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		// get a real token, this can be put in SubTest setup phase
		rs, e := di.AuthClient.PasswordLogin(ctx, seclient.WithCredentials("superuser", "superuser"))
		g.Expect(e).To(Succeed(), "initial access token request should be success")

		// Authentication is passed along to the controller in MockedServer mode.
		// Note that this would override mocked security context in setup func, so we need to re-do it
		ctx = sectest.ContextWithSecurity(ctx, sectest.MockedAuthentication(func(d *sectest.SecurityDetailsMock) {
			d.AccessToken = rs.Token.Value()
			d.Permissions = utils.NewStringSet("DUMMY_PERMISSION")
		}))

		// Do test as usual
		req := webtest.NewRequest(ctx, http.MethodPost, "/remote", strings.NewReader(`{"user": "admin"}`),
			webtest.ContentType("application/json"),
		)
		resp := webtest.MustExec(ctx, req).Response
		assertHttpResponse(t, g, resp, "admin")
	}
}

// SubTestUnitTestWithWithoutSystemAccount
// Calling service directly also requires real access token to work, so same setup
// as SubTestMockedServerWithoutSystemAccount except that permission mocking is not required
func SubTestUnitTestWithWithoutSystemAccount(di *TestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		// get a real token, this can be put in SubTest setup phase
		rs, e := di.AuthClient.PasswordLogin(ctx, seclient.WithCredentials("superuser", "superuser"))
		g.Expect(e).To(Succeed(), "initial access token request should be success")

		// Authentication is passed along to the controller in MockedServer mode.
		ctx = sectest.ContextWithSecurity(ctx, sectest.MockedAuthentication(func(d *sectest.SecurityDetailsMock) {
			d.AccessToken = rs.Token.Value()
		}))

		// Do test as usual
		ret, e := di.Service.CallRemoteWithoutSystemAccount(ctx, "admin")
		g.Expect(e).To(Succeed(), "functions that calling remote service should not fail")
		assertResult(t, g, ret, "admin")
	}
}

// SubTestMockedServerWithCurrentContext
// Because current context is used as-is in service, same configuration as SubTestMockedServerWithoutSystemAccount
func SubTestMockedServerWithCurrentContext(di *TestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		// get a real token, this can be put in SubTest setup phase
		rs, e := di.AuthClient.PasswordLogin(ctx, seclient.WithCredentials("superuser", "superuser"))
		g.Expect(e).To(Succeed(), "initial access token request should be success")

		// Authentication is passed along to the controller in MockedServer mode.
		// Note that this would override mocked security context in setup func, so we need to re-do it
		ctx = sectest.ContextWithSecurity(ctx, sectest.MockedAuthentication(func(d *sectest.SecurityDetailsMock) {
			d.AccessToken = rs.Token.Value()
			d.Permissions = utils.NewStringSet("DUMMY_PERMISSION")
		}))

		// Do test as usual
		req := webtest.NewRequest(ctx, http.MethodGet, "/remote", nil)
		resp := webtest.MustExec(ctx, req).Response
		assertHttpResponse(t, g, resp, "superuser")
	}
}

// SubTestUnitTestWithCurrentContext
// Calling service directly also requires real access token to work, so same setup
// as SubTestMockedServerWithCurrentContext except that permission mocking is not required
func SubTestUnitTestWithCurrentContext(di *TestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		// get a real token, this can be put in SubTest setup phase
		rs, e := di.AuthClient.PasswordLogin(ctx, seclient.WithCredentials("superuser", "superuser"))
		g.Expect(e).To(Succeed(), "initial access token request should be success")

		// Authentication is passed along to the controller in MockedServer mode.
		ctx = sectest.ContextWithSecurity(ctx, sectest.MockedAuthentication(func(d *sectest.SecurityDetailsMock) {
			d.AccessToken = rs.Token.Value()
		}))

		// Do test as usual
		ret, e := di.Service.CallRemoteWithCurrentContext(ctx)
		g.Expect(e).To(Succeed(), "functions that calling remote service should not fail")
		assertResult(t, g, ret, "superuser")
	}
}

/*************************
	internal
 *************************/

func assertHttpResponse(t *testing.T, g *gomega.WithT, resp *http.Response, expectedUser string) {
	g.Expect(resp).To(Not(BeNil()), "response cannot be nil")
	g.Expect(resp.StatusCode).To(Equal(http.StatusOK), "response should be 200")

	var body map[string]interface{}
	e := json.NewDecoder(resp.Body).Decode(&body)
	g.Expect(e).To(Succeed(), "response body should be JSON")
	assertResult(t, g, body, expectedUser)
}

func assertResult(_ *testing.T, g *gomega.WithT, i interface{}, expectedUser string) {
	g.Expect(i).To(Not(BeNil()), "functions that calling remote service should have proper response")
	g.Expect(i).To(BeAssignableToTypeOf(map[string]interface{}{}), "service should return a map")
	m := i.(map[string]interface{})
	g.Expect(m).To(HaveKeyWithValue("username", expectedUser), "body should contains correct username")
}
