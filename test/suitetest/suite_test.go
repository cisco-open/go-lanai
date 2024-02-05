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

package suitetest

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"github.com/onsi/gomega"
	"testing"
)

var counter = &testHookCounter{}

type testHookCounter struct {
	setupCount              int
	pkgSetupCount           int
	pkgOrderedSetupCount    int
	subSetupCount           int
	teardownCount           int
	pkgTeardownCount        int
	pkgOrderedTeardownCount int
	subTeardownCount        int
}

func (c *testHookCounter) mainSetup() error {
	c.pkgSetupCount ++
	return nil
}

func (c *testHookCounter) mainTeardown() error {
	c.pkgTeardownCount ++
	return nil
}

func (c *testHookCounter) mainOrderedSetup() error {
	c.pkgOrderedSetupCount++
	return nil
}

func (c *testHookCounter) mainOrderedTeardown() error {
	c.pkgOrderedTeardownCount++
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
		SetupWithOrder(10, counter.mainOrderedSetup), TeardownWithOrder(10, counter.mainOrderedTeardown),
		TestSetup(counter.setup), TestTeardown(counter.teardown),
		WithOptions(
			TestOptions(test.SubTestSetup(counter.subSetup)),
			TestOptions(test.SubTestTeardown(counter.subTeardown)),
		),
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
	g.Expect(counter.pkgOrderedSetupCount).To(gomega.Equal(1), "Suite setup with order should be invoked once for entire package")
	g.Expect(counter.pkgTeardownCount).To(gomega.Equal(0), "Suite teardown should be invoked once for entire package")
	g.Expect(counter.pkgOrderedTeardownCount).To(gomega.Equal(0), "Suite teardown with order should be invoked once for entire package")

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
