# Test

A significant portion of the complexity in an application comes from the integration of the components. While unit tests
are great for testing the code within a component, integration tests are necessary for testing the integration between
them. It's important to be able to perform integration testing without requiring deployment to your application 
server or connecting to other infrastructure. Doing so lets you test such things as the wiring of your components and database queries. 
It is a great way to increase your test coverage without having to mock all the dependencies.

The [example database service](../examples/database/pkg/controller/v1/examplefriends_test.go) has examples of writing integration
tests for an application using some of the topics described in this document.

## Extending Go Testing
In order to support integration testing, we first have to enhance the go testing framework to support things such as:
1. Using different test runner.
2. Being able to run sub tests.
3. Being able to run before and after hooks for tests and sub tests.

This is achieved by introducing the ```test.RunTest``` function as the entry point for any test. 
The ```opts``` parameter allows the test writer to specify the test runner, before and after hooks, and sub tests.
```go
// RunTest is the entry point of any Test...().
// It takes any context, and run sub tests according to provided Options
func RunTest(ctx context.Context, t *testing.T, opts ...Options) {
```

A test written using this function will look like this:
```go
func TestWithSubTests(t *testing.T) {
	test.RunTest(context.Background(), t,
		test.GomegaSubTest(SubTestAlwaysSucceed(), "SuccessfulTest-1"),
	)
}
```

By default, the test runner is the ```unitTestRunner```, which simply runs the test as a conventional go test. 
In the following sections, we will introduce packages that provides other types of test runner and options to facilitate
integration testing.

## apptest

The ```apptest``` package provides the options to run integration test. In particular, the following function is the 
entry point option for integration tests

```go
// Bootstrap is an entrypoint test.Options that indicates all sub tests should be run within the scope of
// an slim version of bootstrap.App
func Bootstrap() test.Options {
```

This option configures the sub tests to be run using a ```NewFxTestRunner```. 
The ```NewFxTestRunner``` is analogous to the ```bootstrap.NewApp``` method which is used to start an application.
It creates a ```bootstrap.Bootstrapper``` just like the one used in the main application. Instead of kicking off 
the application's long-running processes, this test runner runs the test in cli mode. This allows
the component to be wired in the same way as the main application, but without starting the application itself.

When writing the main application, modules are registered through the ```Use()``` method. In the test, the same module can be
registered with the ```apptest.WithModules``` method. The ```appconfig``` module is registered by the test runner itself. 
This gives the test writer a convenient way to provide configuration properties for the test.

To inject components for tests, the ```apptest.WithFxOptions``` method can be used.

The [examples](/apptest/examples) directory contains examples of using the ```apptest``` package to wire components for integration 
testing.

## webtest

Writing web application is one of the main use cases of the go-lanai framework. In normal application execution, the ```web```
module will start the web engine to listen for incoming requests in one of its ```onStart``` hooks and blocks the application from
exiting. However, this may not be suitable for testing. Instead of registering the ```web``` module, use ```webtest.WithMockServer``` 
or ```webtest.WithRealServer``` to enable the web engine in tests. The difference between these two modes are:

`webtest.WithMockServer`:
- Does not start a real HTTP server, therefore does not require network resources such as available port, firewall settings, etc.
- Mostly used for testing HTTP server-side implementation. (e.g. Controller, Middleware, etc.).
- Usually works together with `webtest.NewRequest()` and `webtest.MustExec()` to create and execute test requests directly on the web engine.

`webtest.WithRealServer`:
- Create real HTTP server, therefore requires network resources. This mode should be used with caution, since such resources may not be consistent on all environments where the test might be run.
- Suitable for mocking server-side implementation.
- Mostly used for testing client-side code that requires real HTTP interactions with another server. (e.g. http client, websocket client, etc.)
- In most cases, depending on the test purpose, `ittest.WithHttpPlayback` should be used instead as long as the remote server is accessible at time of development.
- `webtest.NewRequest()` and `webtest.MustExec()` is optional in this mode. What it really does is to extract the random port automatically, which is also available via `webtest.CurrentPort(ctx)`

The [examples](/webtest/examples) directory contains examples of ```webtest``` in action.

## dbtest

Writing application that persist data to a database is another common use case. In real application, The ```data``` and 
```postgresql``` or ```cockroach``` package are used to enable the modules that provides connection to the database. But in testing, a database
connection is not always available, such as when running test in CI/CD pipeline.

The ```dbtest.WithDBPlayback``` replaces that and provides a replacement database connection with the ability to record and playback
the database queries thanks to the [copyist](https://github.com/cockroachdb/copyist) library.

When writing tests, developer is expected to write the test against a real database. This is called record mode. 
The queries and results are recorded and saved to a file. Once the test is recorded, it can be run in playback mode.
In this mode, no connection to the database is required. If the queries generated by the code matches the recorded queries,
the recorded result will be returned. This allows the test to verify correctness without connecting to the database. 

The principal behind this approach is to detect change from the code under test. If the interaction with the database is not modified, the 
test will pass in playback mode. If the interaction with database was modified, it will generate different queries then what was recorded, 
and the test will fail. This could stem from either a bug, or an intentional change that requires the queries to be re-recorded. 
At this point, re-run the test in record mode. If the test passes, then the code is still correct and the new recording can be committed. 
If the test fails, then a bug is discovered.

By default, `dbtest.WithDBPlayback` will connect to a database using the following connection parameters, which assumes 
a CockroachDB instance running on local host. 
 
- Host:     "127.0.0.1",
- Port:     26257, 
- Username: "root",

**Note: this option will not automatically create the database. The database must be created before running the test in record mode.**

Record mode can be enabled at the package level using

```go
func TestMain(m *testing.M) {
	suitetest.RunTests(m,
		dbtest.EnableDBRecordMode(),
	)
}
```

or by specifying the record flag in ```go test```

```bash
go test -p 1  github.com/cisco-open/go-lanai/pkg/... --record
```

or by specifying the record flag when using ```make```

```bash
make test ARGS="--record -p 1"
``` 

Note that the -p flag is important, because ```go test``` runs tests in packages in parallel. This can cause queries from
different packages to be interleaved. As a result, the recorded queries will be different from the expected queries. Using the ```-p``` flag
disables this behavior and runs the tests sequentially. See ```go help build``` for more information about this flag.

See the [examples](/dbtest/examples) directory for examples of using the ```dbtest``` package. 

## suitetest

The ```suitetest``` package gives test writer the option to provide setup in the ```TestMain``` method. 
It is used to provide setups that are needed for all the tests in the same package. For example, the ```dbtest.EnableDBRecordMode()``` option 
enables DB record mode, and the ```embedded.Redis()``` option starts an embedded Redis server.

```suittest.RunTests``` takes either ```PackageHook``` option or ```TestOption```. ```PackageHook``` option like ```dbtest.EnableDBRecordMode()```
is executed once per package. ```TestOption``` like ```embedded.Redis()``` is executed for each top level test.

## sectest

In most cases an application will have middleware that creates a security context for the incoming request. Application code
may have logic that depends on the security context. In order to test this code, the security context must be mocked. 
The ```sectest``` package provides options to facilitate tests that require security context.

### Mock Current Security Context 
If security middleware is installed in the application, the security context of an incoming request is passed along in 
```context.Context```.

For example, a controller method with the following signature can expect to extract the security context from the ```context.Context``` parameter.

```go
func (c *ExampleFriendsController) GetItems(ctx context.Context) (int, interface{}, error) {
```

To test this method, use the following method to mock the `ctx` parameter.
```go 
func ContextWithSecurity(ctx context.Context, opts ...SecurityContextOptions) context.Context {
```

The test can use the returned `context.Context` directly with the ```GetItems``` method to test the logic that depends on 
the security context.

### Mock Security Middleware
In some cases, it is necessary to mock the security context for the incoming request. For example the test code tests the entire
path from HTTP request to HTTP response, or there are code in the middlewares before the controller that needs to be tested. 
The ```sectest.WithMockedMiddleware``` option provides a way to do this.

The ```sectest.WithMockedMiddleware``` option does this by enabling the ```security.Module``` and using the mechanism provided by this module 
to add a mocked middleware. (Internally this is done by adding the mocked middleware feature to ```security.WebSecurity```.) 
It allows the test writer to configure the behaviour of the mocked middleware by passing options to the ```sectest.WithMockedMiddleware``` method.

When used with ```webtest.WithMockedServer``` the default behaviour of this option is the same as the default behaviour of ```webtest.WithMockedServer```, 
which is to use the ```security.Authentication``` from the request's context. This is redundant because request's context 
is automatically linked with ```gin.Context``` when using ```webtest.WithMockedServer```, therefore use this option in the
presence of ```webtest.WithMockedServer``` only if there is a need to mock the security context dynamically based on incoming request.
In which case, the test will need to provide a custom ```sectest.MWMocker```.

When used with ```webtest.WithRealServer```, a custom ```sectest.MWMocker``` is required. It can be provided by:
- Using the ```sectest.MWCustomMocker``` option
- Providing a ```sectest.MWMocker``` using uber/fx
- Providing a ```security.Configurer``` with ```sectest.NewMockedMW```:
```go
func realServerSecConfigurer(ws security.WebSecurity) {
 ws.Route(matcher.AnyRoute()).
 With(sectest.NewMockedMW().
    Mocker(sectest.MWMockFunc(realServerMockFunc)),
 )
}
```

### Mock Scope
In some applications, code needs to be executed in a different scope instead of the current security context. go-lanai's
```scope``` package provides a way to switch the current execution to a different security context. In order to test application code that
utilizes this package, ```sectest``` provides a way to mock security scopes, so that the code under test can switch to 
them, and the test can verify the result matches expectations. This can be done with the `sectest.WithMockedScopes` option.
The test writer can provide mocked accounts, tenants and integration clients using a yaml file.

See the [examples](sectest/examples) directory for examples of using this option.

## ittest
Some application needs to interact with other services via HTTP. The ```httpclient``` package provides a way to do this.
In order to test the application code that uses this package, the ```ittest``` package provides a way to record and playback
the HTTP requests and responses. In addition to this, the ```ittest``` package works for any situation where a `http.Client`
is used to make HTTP requests. The principal behind this approach is the same as the one used in the ```dbtest``` package.
Running tests in playback mode will detect changes in the interaction between the client and server. A failed test indicates
there was change in the underlying code, which could be either a bug, or that the interaction needs to be re-recorded due to
an intentional change.

```ittest.WithHttpPlayback```:

This option enables the HTTP playback feature by switching the HTTP client to a client whose ```transport``` is a special
```http.RoundTripper``` that is capable of recording and playing back HTTP requests and responses. 

By default, this option is in playback mode. To enable record mode, use one of the following options:
1. Set the `--record-http` flag when running the test from command line using ```go test``` or ```make test``` similar to the 
   ```dbtest``` package.
2. Use the ```ittest.HttpRecordingMode()``` option in ```ittest.WithHttpPlayback``` to enable record mode for that test.
3. Use the ```ittest.PackageHttpRecordingMode``` option in ```TestMain``` to enable record mode for all tests in the package. 

If the application uses a microservice architecture, it may need to interact with other microservices using HTTP. If the
application uses service discovery to look up the target microservice, the test can use ```sdtest.WithMockedSD``` to mock
the service discovery client. This allows the test to control the resolved address of the target microservice and point it
to the target microservice without going through the real service discovery mechanism (e.g. DNS, service registrar). 

```ittest.WithRecordedScopes()```:

One of the special cases of HTTP interaction in a microservice architecture is to call the authorization server to switch 
the security context. Assuming the authorization server is written in go-lanai, this can be done using the ```scope``` package.

One strategy to test code that switch security context is to use the ```sectest.WithMockedScopes``` option. Alternatively, test
can record the interaction with the authorization server using the ```ittest.WithRecordedScopes()``` option in combination with the 
```ittest.WithHttpPlayback``` option. This option replaces ```sectest.WithMockedScopes``` by recording and playing back the
interaction with the authorization server instead of mocking the scopes.

See the [examples](ittest/examples) directory for examples of using the ```ittest``` package.

### Usage of ittest in Other Packages and Scenarios
In general ```ittest``` can be used to record and playback any situation that uses `http.Client`. When ```ittest.WithHttpPlayback```
is present, a ```*recorder.Recorder``` is available for injection. This recorder instance can be used to create a `http.Client` that 
is capable of recording and playback. It can also be used to wrap an existing `http.Client`'s transport so that it's capable of recording
and playback.

```consultest``` and ```opensearchtest``` packages uses this principal to record and playback the HTTP requests
and responses made by consul client and open search client respectively.

## Misc
In addition to these packages, there are other packages that provides test utilities that facilitates testing in go-lanai.
For example, the ```kafkatest``` package provides a way to mock messages for receivers, or to inspect messages from producers.
They work with the integration testing methodology described above. Explore these packages to see how they can help you 
write tests for your application.