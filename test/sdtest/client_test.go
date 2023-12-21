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

package sdtest

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/discovery"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"errors"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"testing"
	"time"
)

const (
	svc1 = "svc1"
	svc2 = "svc2"
)

func TestClientMock(t *testing.T) {
	test.RunTest(context.Background(), t,
		test.GomegaSubTest(SubTestBasicDiscovery(), "TestTestBasicDiscovery"),
		test.GomegaSubTest(SubTestDiscoveryUpdate(), "TestTestDiscoveryUpdate"),
		test.GomegaSubTest(SubTestDiscoveryError(), "TestTestDiscoveryError"),
	)
}

/*************************
	Sub Tests
 *************************/

func SubTestBasicDiscovery() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		client := NewMockClient(ctx)
		var extraTagSet bool
		var healthy int
		client.MockService(svc1, 3, func(inst *discovery.Instance) {
			if !extraTagSet {
				inst.Tags = append(inst.Tags, "extra-tag")
				extraTagSet = true
			}
			if healthy >= 2 {
				inst.Health = discovery.HealthCritical
			} else {
				healthy ++
			}
		})

		// find all
		assertClient(t, g, client, svc1, nil, 2, false)

		// find healthy
		assertClient(t, g, client, svc1, discovery.InstanceIsHealthy(), 2, false)

		// find with tag
		assertClient(t, g, client, svc1,
			discovery.InstanceIsHealthy().And(discovery.InstanceWithTag("Extra-Tag", true)),
			1, false)
	}
}

func SubTestDiscoveryUpdate() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		client := NewMockClient(ctx)
		// update without mock
		count := client.UpdateMockedService(svc1, NthInstance(1), BeCritical())
		g.Expect(count).To(BeZero(), "ClientMock.UpdateMockedService() should return 0 before mocks")
		assertClient(t, g, client, svc1, nil, 0, false)

		// initial mock
		client.MockService(svc1, 3)
		assertClient(t, g, client, svc1, nil, 3, false)

		// one down
		client.UpdateMockedService(svc1, NthInstance(1), BeCritical())
		assertClient(t, g, client, svc1, discovery.InstanceIsHealthy(), 2, false)

		// all back up
		client.UpdateMockedService(svc1, AnyInstance(), BeHealthy())
		assertClient(t, g, client, svc1, discovery.InstanceIsHealthy(), 3, false)

		// extra tag
		client.UpdateMockedService(svc1, InstanceAfterN(1), WithExtraTag("extra-tag"))
		assertClient(t, g, client, svc1,
			discovery.InstanceIsHealthy().And(discovery.InstanceWithTag("Extra-Tag", true)),
			1, false)

		// with Meta
		client.UpdateMockedService(svc1, InstanceAfterN(0), WithMeta("meta", "value"))
		assertClient(t, g, client, svc1,
			discovery.InstanceIsHealthy().And(discovery.InstanceWithMetaKV("meta", "value")),
			2, false)
	}
}

func SubTestDiscoveryError() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {

		client := NewMockClient(ctx)
		mockedErr := errors.New("oops")

		// mock error
		client.MockError(svc2, mockedErr, time.Now())
		assertClient(t, g, client, svc2, nil, 0, true)

		// reset error by mock again
		client.MockService(svc2, 2)
		assertClient(t, g, client, svc2, nil, 2, false)

		// mock error again
		client.MockError(svc2, mockedErr, time.Now())
		assertClient(t, g, client, svc2, nil, 0, true)

		// reset error by update
		client.UpdateMockedService(svc2, AnyInstance(), func(inst *discovery.Instance) {
			inst.Tags = append(inst.Tags, "some-tag")
		})
		assertClient(t, g, client, svc2, nil, 2, false)
	}
}

/*************************
	Helpers
 *************************/

func assertClient(t *testing.T, g *gomega.WithT, client *ClientMock, svcName string, matcher discovery.InstanceMatcher, expectedCount int, expectErr bool) {
	instancer, e := client.Instancer(svcName)
	g.Expect(e).To(Succeed(), "Instancer shouldn't return error")
	g.Expect(instancer).To(BeAssignableToTypeOf(NewMockInstancer(nil, "")), "Instancer shouldn't return nil")
	g.Expect(instancer.ServiceName()).To(BeEquivalentTo(svcName), "Instancer.ServiceName() should be correct")

	// start and stop
	instancer.Stop()
	g.Expect(instancer.(*InstancerMock).Started).To(BeFalse(), "Instancer should be stopped")
	instancer.Start(client.Context())
	g.Expect(instancer.(*InstancerMock).Started).To(BeTrue(), "Instancer should be started")

	// following should be noop, check there is no panic
	instancer.RegisterCallback("whatever", func(instancer discovery.Instancer) {})
	instancer.DeregisterCallback("whatever")

	// check mocked instances and service
	if insts, e := instancer.Instances(matcher); expectErr {
		g.Expect(e).To(HaveOccurred(), "Instancer.Instances() should return error")
	} else {
		g.Expect(e).To(Succeed(), "Instancer.Instances() shouldn't return error")
		g.Expect(insts).To(HaveLen(expectedCount), "Instancer.Instances() should return %d instances that %v", expectedCount, matcher)
	}

	if svc := instancer.Service(); expectErr {
		g.Expect(svc.Err).To(HaveOccurred(), "Instancer.Service() should return error")
		g.Expect(svc.FirstErrAt).To(Not(BeZero()), "Instancer.Service() should return non-zero error time")
	} else {
		g.Expect(svc.Err).To(Succeed(), "Instancer.Service() shouldn't return error")
		g.Expect(len(svc.Insts)).To(BeNumerically(">=", expectedCount), "Instancer.Service() should return at least %d instances", expectedCount)
	}
}
