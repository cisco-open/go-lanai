# Bootstrap

When writing a go-lanai application, the developer selects the go-lanai module they want to use. Bootstrap refers to the process of
how the application instantiate the components needed by the modules and wires them together.

Under the hood, the bootstrapper keeps a registry of modules that's enabled in this application.
A module is implemented as a group of ```fx.Provide``` and ```fx.Invoke``` options.
When the app starts, all module added ```fx.provide``` and ```fx.invoke``` options are sorted and executed.

In go-lanai's module packages, you'll usually see a function like this:

```go
package init

var Module = &bootstrap.Module{
	Name: "web",
	Precedence: web.MinWebPrecedence,
	PriorityOptions: []fx.Option{
		appconfig.FxEmbeddedDefaults(defaultConfigFS),
		fx.Provide(
			web.BindServerProperties,
			web.NewEngine,
			web.NewRegistrar),
		fx.Invoke(setup),
	},
}

// Use Allow service to include this module in main()
func Use() {
	bootstrap.Register(Module)
	bootstrap.Register(cors.Module)
}
```

This Use() method registers the module with the bootstrapper. The application
code just needs to call the ```Use()``` function to indicate this module should be activated in the application.