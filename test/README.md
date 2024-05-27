# Test

A significant portion of the complexity in an application comes from the integration of the components. While unit tests
are great for testing the complexity within a component, integration tests are necessary for testing the integration between
components. It's important to be able to perform integration testing without requiring deployment to your application 
server or connecting to other infrastructure. Doing so lets you test such things as the wiring of your components and database queries. 
It is a great way to increase your test coverage without having to mock all the dependencies.

## Extending Go Testing
In order to support integration testing, we first have to enhance the go testing framework to support things such as:
1. Using different test runner.
2. Being able to run sub tests.
3. Being able to run before and after hooks for tests and sub tests.

This is achieved by introducing the ```RunTest``` function as the entry point for any test. 
The ```opts``` parameter allows the test writer to specify the test runner, before and after hooks, and sub tests.
```go
// RunTest is the entry point of any Test...().
// It takes any context, and run sub tests according to provided Options
func RunTest(ctx context.Context, t *testing.T, opts ...Options) {
```

A test written using this function will look like this:
```go
func TestWithSubTests(t *testing.T) {
	RunTest(context.Background(), t,
		GomegaSubTest(SubTestAlwaysSucceed(), "SuccessfulTest-1"),
	)
}
```

By default, the test runner is the ```unitTestRunner```, which simply runs the test as a normal go test. 
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
The ```NewFxTestRunner``` is equivalent to the ```bootstrap.NewApp``` method which is used to start an application.
It creates a ```bootstrap.Bootstrapper``` just like the one used in the main application. Instead of kicking off 
the application's long-running processes, the test runner runs the test in cli mode. This allows
the component to be wired in the same way as the main application, but without starting the application itself.

When writing the main application, modules are registered through the ```Use()``` method. In the test, the same module can be
registered with the ```apptest.WithModules``` method. The ```appconfig``` modules is registered by the test runner itself. 
This gives the test writer a convenient way to provide configuration properties for the test.

To inject components for tests, the ```apptest.WithFxOptions``` method can be used.

The [examples](/apptest/examples) directory contains examples of using the ```apptest``` package to wire components for integration 
testing.

## webtest

Writing web application is one of the main use cases of the go-lanai framework. In normal application execution, the ```web```
module will start the web engine to listen for incoming requests in one of its ```onStart``` hooks and blocks the application from
exiting. To facilitate testing web applications, the ```webtest``` package provides utility methods such as ```webtest.NewRequest``` 
and ```webtest.Exec``` to allow the test writer to create request and test their execution against the web engine.

Instead of registering the ```web``` module, use ```webtest.WithRealServer``` or ```webtest.WithMockServer``` to enable
the web engine for testing. The term "RealServer" is used to mean that the web engine is created the same way as the ```web```
module. In this mode, the web engine is started and listens on a random port. Request is send to the web engine using a 
real http connection.

The term "MockServer" is used to mean that the web engine is created in the ```webtest``` package itself. 
In this mode, the web engine is not started. Instead, the request is created using `httptest.NewRecorder()` and pass to 
the web engine directly.

The [examples](/webtest/examples) directory contains examples of ```webtest``` in action.

## dbtest

Writing application that persist data to a database is another common use case. In real application, The ```data``` and 
```postgres``` package are used to enable the modules that provides connection to the database. But in testing, a database
connection is not always available, such as when running test in CI/CD pipeline.

The `dbtest.WithDBPlayback` replaces that and provides a replacement database connection with the ability to record and playback
the database queries thanks to the [copyist](https://github.com/cockroachdb/copyist) library.

When writing tests, developer is expected to write the test against a real database. This is called record mode. 
The queries and results are recorded and saved to a file. Once the test is written, it can be run in playback mode.
In this mode, no connection to the database is required. If the queries generated by the test code matches the recorded queries,
the recorded result will be returned. This allows the test to be run without a connection to the database. The tests also
executes at faster speed in this mode.

Record mode can be enabled at the package level using

```go
func TestMain(m *testing.M) {
	suitetest.RunTests(m,
		dbtest.EnableDBRecordMode(),
	)
}
```

or by specifying the record flag to ```go test```

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

## Suitetest

The ```suitetest``` package gives test writer the option to provide setup in the ```TestMain``` method. 
It is used to provide setups that are needed for all the tests in the same package. For example, the ```dbtest.EnableDBRecordMode()``` option 
enables DB record mode, and the ```embedded.Redis()``` option starts an embedded Redis server.

```suittest.RunTests``` takes either ```PackageHook``` option or ```TestOption```. ```PackageHook``` option like ```dbtest.EnableDBRecordMode()```
is executed once per package. ```TestOption``` like ```embedded.Redis()``` is executed for each top level test.