package suitetest

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"github.com/onsi/gomega"
	"testing"
)

var counter = &testHookCounter{}

type testHookCounter struct {
	setupCount       int
	pkgSetupCount    int
	subSetupCount    int
	teardownCount    int
	pkgTeardownCount int
	subTeardownCount int
}

func (c *testHookCounter) mainSetup() error {
	c.pkgSetupCount ++
	return nil
}

func (c *testHookCounter) mainTeardown() error {
	c.pkgTeardownCount ++
	return nil
}

func (c *testHookCounter) setup(ctx context.Context, _ *testing.T) (context.Context, error) {
	c.setupCount ++
	return ctx, nil
}

func (c *testHookCounter) teardown(_ context.Context, _ *testing.T) error {
	c.teardownCount ++
	return nil
}

func (c *testHookCounter) subSetup(ctx context.Context, _ *testing.T) (context.Context, error) {
	c.subSetupCount ++
	return ctx, nil
}

func (c *testHookCounter) subTeardown(_ context.Context, _ *testing.T) error {
	c.subTeardownCount ++
	return nil
}

/*************************
	Test Main Setup
 *************************/

func TestMain(m *testing.M) {
	RunTests(m,
		Setup(counter.mainSetup), Teardown(counter.mainTeardown),
		TestSetup(counter.setup), TestTeardown(counter.teardown),
		TestOptions(test.SubTestSetup(counter.subSetup)),
		TestOptions(test.SubTestTeardown(counter.subTeardown)),
	)
}

/*************************
	Test Cases
 *************************/
func TestSuiteHookInvocations(t *testing.T) {
	g := gomega.NewWithT(t)
	g.Expect(counter.pkgSetupCount).To(gomega.Equal(1), "Suite setup should be invoked once for entire package")
	// our test is not finished, package teardown count should be 0
	g.Expect(counter.pkgTeardownCount).To(gomega.Equal(0), "Suite teardown should be invoked only once for entire package after all tests finished")
}

func TestAllHookInvocations(t *testing.T) {
	localCounter := &testHookCounter{}
	test.RunTest(context.Background(), t,
		test.Setup(localCounter.setup),
		test.Teardown(localCounter.teardown),
		test.SubTestSetup(localCounter.subSetup),
		test.SubTestTeardown(localCounter.subTeardown),
		test.AnonymousSubTest(DummySubTest()),
		test.AnonymousSubTest(DummySubTest()),
		test.AnonymousSubTest(DummySubTest()),
	)

	g := gomega.NewWithT(t)
	g.Expect(counter.pkgSetupCount).To(gomega.Equal(1), "Suite setup should be invoked once for entire package")
	g.Expect(counter.pkgTeardownCount).To(gomega.Equal(0), "Suite teardown should be invoked once for entire package")

	g.Expect(counter.setupCount).To(gomega.Equal(1), "Suite's test setup should be invoked once per test")
	g.Expect(counter.teardownCount).To(gomega.Equal(1), "Suite's test teardown should be invoked once per test")
	g.Expect(localCounter.setupCount).To(gomega.Equal(1), "Local test setup should be invoked once per test")
	g.Expect(localCounter.teardownCount).To(gomega.Equal(1), "Local test teardown should be invoked once per test")

	g.Expect(counter.subSetupCount).To(gomega.Equal(3), "Local SubTest setup should invoked once per sub test")
	g.Expect(counter.subTeardownCount).To(gomega.Equal(3), "Local SubTest teardown should invoked once per sub test")
	g.Expect(localCounter.subSetupCount).To(gomega.Equal(3), "Local SubTest setup should invoked once per sub test")
	g.Expect(localCounter.subTeardownCount).To(gomega.Equal(3), "Local SubTest teardown should invoked once per sub test")
}

/*************************
	Sub-Test Cases
 *************************/
func DummySubTest() test.SubTestFunc {
	return func(ctx context.Context, t *testing.T) {

	}
}

/*************************
	Helpers
 *************************/
