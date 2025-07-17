package cmdtest

import "github.com/cisco-open/go-lanai/test"

func WithDryRun(handler TestDryRunHandler) test.Options {
	return test.WithOptions(
		test.SubTestSetup(SetupResetPackageVars()),
		test.SubTestSetup(SetupDryRun(handler)),
	)
}
