package actuatortest

type ActuatorOptions func(opt *ActuatorOption)
type ActuatorOption struct {
	// Default to false. When set true, the default health, info and env endpoints are not initialized
	DisableAllEndpoints    bool
	// Default to true. When set to false, the default authentication is installed.
	// Depending on the defualt authentication (currently tokenauth), more dependencies might be needed
	DisableDefaultAuthentication bool
}

// DisableAllEndpoints is an ActuatorOptions that disable all endpoints in test.
// Any endpoint need to be installed manually via apptest.WithModules(...)
func DisableAllEndpoints() ActuatorOptions {
	return func(opt *ActuatorOption) {
		opt.DisableAllEndpoints = true
	}
}

