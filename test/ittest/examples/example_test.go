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
	"time"
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
// Behavior based tests that verify Controllers end-to-end behavior by sending requests to a Mocked Server.
//
// Please see "example_service.go" for the test subject setup:
// - ExampleController is properly mapped and protected by permissions
// - ExampleController uses ExampleService's corresponding method based on the input
// - ExampleService perform security context switching and make HTTP call to another service (IDM in this case)
// This test style mimic common microservice tests tasks
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
		// Note: this function accept many options. See ittest/httpvcr.go for more details
		ittest.WithHttpPlayback(t), // Enable recording mode to use real service for any HTTP interaction.
		// This should be enabled during development and turned off before checking in the code
		//ittest.HttpRecordingMode(),

		// Because the test subjects (ExampleController, ExampleService) uses service discovery and scopes,
		// They need to be configured properly for HTTP recorder to work

		// Note: We use ittest.WithRecordedScopes instead of sectest.WithMockedScopes, because switching context
		// 		 in recording mode requires real access token in each scope. The HTTP interactions during context
		// 		 switching is also recorded
		ittest.WithRecordedScopes(),
		// httpclient.Module requires a discovery.Client to work.
		// See `sdtest` for more examples. The SD mocking is defined in testdata/application-test.yml
		sdtest.WithMockedSD(sdtest.DefinitionWithPrefix("mocks.sd")),

		// Because remote access require current context to be authentic during recording mode,
		// We need this configuration to pass along the security context from the test context to Controller's context.
		// Note: 	Since gin-gonic v1.8.0+, this setup is not required anymore for webtest.WithMockedServer. Values in
		//			request's context is automatically linked with gin.Context.
		//sectest.WithMockedMiddleware(),

		// Controller requires permissions, we need to mock it for any test that uses webtest.
		// Note: this only works with webtest.WithMockedServer() and sectest.WithMockedMiddleware() together.
		test.SubTestSetup(MockPermissions()),

		// Test order is important, unless ittest.DisableHttpRecordOrdering option is used
		test.GomegaSubTest(SubTestMockedServerWithSystemAccount(), "TestMockedServerWithSystemAccount"),
		test.GomegaSubTest(SubTestMockedServerWithoutSystemAccount(&di), "TestMockedServerWithoutSystemAccount"),
		test.GomegaSubTest(SubTestMockedServerWithCurrentContext(&di), "TestMockedServerWithCurrentContext"),
	)
}

// TestExampleUnitTestWithSecurity
// Unit tests that verify business logic components (ExampleService) that internally perform security context switching
// and remote HTTP call.
//
// Please see "example_service.go" for the test subject setup:
// - ExampleService perform security context switching and make HTTP call to another service (IDM in this case)
// This test style is usually suitable for standalone components that not directly invoked in response of incoming HTTP requests.
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
		ittest.WithHttpPlayback(t), // Enable recording mode to use real service for any HTTP interaction.
		// This should be enabled during development and turned off before checking in the code
		//ittest.HttpRecordingMode(),

		// Because the test subjects (ExampleService) uses service discovery and scopes,
		// They need to be configured properly for HTTP recorder to work

		// Note: We use ittest.WithRecordedScopes instead of sectest.WithMockedScopes, because switching context
		// 		 in recording mode requires real access token in each scope. The HTTP interactions during context
		// 		 switching is also recorded
		ittest.WithRecordedScopes(),
		// httpclient.Module requires a discovery.Client to work.
		// See `sdtest` for more examples. The SD mocking is defined in testdata/application-test.yml
		sdtest.WithMockedSD(sdtest.DefinitionWithPrefix("mocks.sd")),

		// Test order is important, unless ittest.DisableHttpRecordOrdering option is used
		test.GomegaSubTest(SubTestUnitTestWithSystemAccount(&di), "TestUnitTestWithSystemAccount"),
		test.GomegaSubTest(SubTestUnitTestWithWithoutSystemAccount(&di), "TestUnitTestWithWithoutSystemAccount"),
		test.GomegaSubTest(SubTestUnitTestWithCurrentContext(&di), "TestUnitTestWithCurrentContext"),
	)
}

type AnotherTestDI struct {
	fx.In
	HttpClient httpclient.Client
}

// TestExampleCustomRequestMatching
// This example demonstrate some custom request matching options.
// By default, requests need to be replayed in exact order
func TestExampleCustomRequestMatching(t *testing.T) {
	var di AnotherTestDI
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		// Setup your test, includes modules you need, provide test subjects and other mocks
		apptest.WithModules(httpclient.Module),
		apptest.WithDI(&di),

		// Tell test framework to use recorded HTTP interaction.
		// Note: this function accept may options. See ittest/httpvcr.go for more details
		ittest.WithHttpPlayback(t,
			// Enable recording mode to use real service for any HTTP interaction.
			// This should be enabled during development and turned off before checking in the code
			//ittest.HttpRecordingMode(),

			// Disable Host matching when replaying. Useful when remote server has random port
			ittest.HttpRecordIgnoreHost(),

			// Fuzzy request matching for entire test.
			// Per sub-test customization is also possible using ittest.AdditionalMatcherOptions
			// Use case: recorded HTTP interactions may contain temporal/random data in headers/queries/body that would
			// 			 cause replaying difficult. We can customize our request matching to ignore those values.
			// Note: the request still need to contain those headers/queries/fields to be matched. Only value comparison is disabled.
			ittest.HttpRecordMatching(
				// When matching request, values of specified keys in form data (queries and x-form-urlencoded body) are ignored.
				ittest.FuzzyHeaders("X-Date"),
				// When matching request, values of specified keys in form data (queries and x-form-urlencoded body) are ignored.
				ittest.FuzzyForm("time", "random"),
				// When matching request, values of specified JSONPath in JSON body are ignored.
				// JSONPath Syntax: https://goessner.net/articles/JsonPath/
				ittest.FuzzyJsonPaths("$..time"),
			),
		),

		// httpclient.Module requires a discover.Client to work.
		sdtest.WithMockedSD(sdtest.DefinitionWithPrefix("mocks.sd")),

		// Test order is important, unless ittest.DisableHttpRecordOrdering option is used
		test.GomegaSubTest(SubTestCustomRequestMatching(&di), "CustomRequestMatching"),
		test.GomegaSubTest(SubTestPerSubTestCustomRequestMatching(&di), "PerSubTestCustomRequestMatching"),
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

func SubTestCustomRequestMatching(di *AnotherTestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		client, e := di.HttpClient.WithBaseUrl("http://127.0.0.1:8081/europa")
		g.Expect(e).To(Succeed(), "client with base URL should be available")

		now := time.Now()
		body := map[string]interface{}{
			"fixed-value": "this will be compared",
			"object": map[string]interface{}{
				"fixed-value": "this will be compared",
				"time":        now.Format(time.RFC3339), // this value will be ignored
			},
		}
		// With following temporal/random values in request, this interaction would normally fail during replay mode.
		// However, with Fuzzy* options, they will match the recorded interaction even the values doesn't match.
		req := httpclient.NewRequest("/public/ping", http.MethodPost,
			httpclient.WithHeader("X-Date", now.Format(time.RFC850)),
			httpclient.WithParam("time", now.Format(time.RFC3339)),
			httpclient.WithParam("random", utils.RandomString(10)),
			httpclient.WithBody(body),
		)
		resp, e := client.Execute(ctx, req, httpclient.JsonBody(&map[string]interface{}{}))
		g.Expect(e).To(Succeed(), "execute request with random values shouldn't fail due to Fuzzy matching")
		g.Expect(resp.StatusCode).To(Equal(http.StatusOK), "response should be 200")
		g.Expect(resp.Body).To(HaveKeyWithValue("message", "pong"), "response should be correct")
	}
}

func SubTestPerSubTestCustomRequestMatching(di *AnotherTestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {

		// Fuzzy request matching can be set for particular sub-test.
		// Any additional matcher options takes effect only for this sub-test.
		// All test level options still in effect
		ittest.AdditionalMatcherOptions(ctx, ittest.FuzzyJsonPaths("$.random"))

		client, e := di.HttpClient.WithBaseUrl("http://127.0.0.1:8081/europa")
		g.Expect(e).To(Succeed(), "client with base URL should be available")

		now := time.Now()
		body := map[string]interface{}{
			"fixed-value": "this will be compared",
			"object": map[string]interface{}{
				"fixed-value": "this will be compared",
				"time":        now.Format(time.RFC3339), // this value will be ignored
			},
			// This would fail with test level matching configuration, but should succeed in this sub-test
			"random": utils.RandomString(20),
		}
		// Any test-level matching options should still in effect
		req := httpclient.NewRequest("/public/ping", http.MethodPost,
			httpclient.WithHeader("X-Date", now.Format(time.RFC850)),
			httpclient.WithParam("time", now.Format(time.RFC3339)),
			httpclient.WithParam("random", utils.RandomString(10)),
			httpclient.WithBody(body),
		)
		resp, e := client.Execute(ctx, req, httpclient.JsonBody(&map[string]interface{}{}))
		g.Expect(e).To(Succeed(), "execute request with random values shouldn't fail due to Fuzzy matching")
		g.Expect(resp.StatusCode).To(Equal(http.StatusOK), "response should be 200")
		g.Expect(resp.Body).To(HaveKeyWithValue("message", "pong"), "response should be correct")
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
