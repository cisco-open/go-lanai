package kafkatest

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/apptest"
	"go.uber.org/fx"
)

// WithMockedBinder returns a test.Options that provides mocked kafka.Binder and a MessageRecorder.
// Tests can wire the MessageRecorder and verify invocation of kafka.Producer
// Note: The main purpose of this test configuration is to fulfill dependency injection and validate kafka.Producer is
//		 invoked as expected. It doesn't validate/invoke any message options such as ValueEncoder or Key, nor does it
//		 respect any binding configuration
func WithMockedBinder() test.Options {
	testOpts := []test.Options{
		apptest.WithFxOptions(
			fx.Provide(provideMockedBinder),
		),
	}
	return test.WithOptions(testOpts...)
}



