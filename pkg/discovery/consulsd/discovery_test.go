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

package consulsd_test

import (
	"context"
	"fmt"
	"github.com/cisco-open/go-lanai/pkg/bootstrap"
	"github.com/cisco-open/go-lanai/pkg/consul"
	"github.com/cisco-open/go-lanai/pkg/discovery"
	"github.com/cisco-open/go-lanai/pkg/discovery/consulsd"
	"github.com/cisco-open/go-lanai/pkg/discovery/consulsd/testdata"
	"github.com/cisco-open/go-lanai/test"
	"github.com/cisco-open/go-lanai/test/apptest"
	"github.com/cisco-open/go-lanai/test/consultest"
	"github.com/cisco-open/go-lanai/test/ittest"
	"github.com/go-kit/kit/sd"
	"github.com/hashicorp/consul/api"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
	"io"
	"reflect"
	"sort"
	"sync"
	"testing"
	"time"
)

const (
	ServiceName1  = "service1"
	ServiceName2  = "service2"
	Service1Port1 = 8011
	Service1Port2 = 8012
	Service1Port3 = 8013
	Service2Port1 = 8021
	//Service2Port2 = 8022
	//Service2Port3 = 8023
)

var MockedServices = map[string][]MockedService{
	ServiceName1: {
		MockedService{
			Name:    ServiceName1,
			Port:    Service1Port1,
			Tags:    []string{"instance=1", "legacy=true"},
			Meta:    map[string]string{"instance": "1", discovery.InstanceMetaKeyVersion: "0.0.1", "legacy": "true"},
			Healthy: true,
		},
		MockedService{
			Name:    ServiceName1,
			Port:    Service1Port2,
			Tags:    []string{"instance=2"},
			Meta:    map[string]string{"instance": "2", discovery.InstanceMetaKeyVersion: "0.0.2"},
			Healthy: true,
		},
		MockedService{
			Name:    ServiceName1,
			Port:    Service1Port3,
			Tags:    []string{"instance=3", "legacy=true"},
			Meta:    map[string]string{"instance": "3", discovery.InstanceMetaKeyVersion: "0.0.1", "legacy": "true"},
			Healthy: false,
		},
	},
	ServiceName2: {
		MockedService{
			Name:    ServiceName2,
			Port:    Service2Port1,
			Tags:    []string{"instance=1"},
			Meta:    map[string]string{"instance": "1"},
			Healthy: true,
		},
	},
}

/*************************
	Tests
 *************************/

type TestDiscoveryDI struct {
	fx.In
	Consul     *consul.Connection
	AppContext *bootstrap.ApplicationContext
	Registrar  discovery.ServiceRegistrar
}

func TestDiscoveryClient(t *testing.T) {
	di := TestDiscoveryDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		consultest.WithHttpPlayback(t,
			//consultest.HttpRecordingMode(),
			// Note: tags may contains build time, should be ignored
			consultest.MoreHTTPVCROptions(
				ittest.HttpRecordMatching(ittest.FuzzyJsonPaths(
					TestRegisterFuzzyJsonPathTags,
					TestRegisterFuzzyJsonPathMeta,
				)),
				ittest.ApplyHttpLatency(),
			),
		),
		apptest.WithBootstrapConfigFS(testdata.TestBootstrapFS),
		apptest.WithConfigFS(testdata.TestApplicationFS),
		apptest.WithFxOptions(fx.Provide(consulsd.NewServiceRegistrar)),
		apptest.WithDI(&di),
		test.SubTestSetup(SetupTestServices(&di)),
		test.SubTestTeardown(TeardownTestServices(&di)),
		test.GomegaSubTest(SubTestInstancerManagement(&di), "TestInstancerManagement"),
		test.GomegaSubTest(SubTestWithDefaultSelector(&di), "TestWithDefaultSelector"),
		test.GomegaSubTest(SubTestWithSelectors(&di), "TestWithSelectors"),
		test.GomegaSubTest(SubTestWithServiceUpdates(&di), "TestWithServiceUpdates"),
		test.GomegaSubTest(SubTestWithGoKitCompatibility(&di), "TestWithGoKitCompatibility"),
	)
}

/*************************
	Sub-Test Cases
 *************************/

func SetupTestServices(di *TestDiscoveryDI) test.SetupFunc {
	return func(ctx context.Context, t *testing.T) (context.Context, error) {
		OrderedDoForEachMockedService(func(reg *consulsd.ServiceRegistration) {
			_ = di.Registrar.Register(ctx, reg)
		})
		return ctx, nil
	}
}

func TeardownTestServices(di *TestDiscoveryDI) test.TeardownFunc {
	return func(ctx context.Context, t *testing.T) error {
		OrderedDoForEachMockedService(func(reg *consulsd.ServiceRegistration) {
			_ = di.Registrar.Deregister(ctx, reg)
		})
		return nil
	}
}

func SubTestInstancerManagement(di *TestDiscoveryDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		client := consulsd.NewDiscoveryClient(ctx, di.Consul)
		g.Expect(client.Context()).To(Equal(ctx), "client's context should be correct")

		// get instancer
		instancer, e := client.Instancer(ServiceName1)
		g.Expect(e).To(Succeed(), "getting instancer should not fail")
		g.Expect(instancer).ToNot(BeNil(), "instancer should not be nil")
		g.Expect(instancer.ServiceName()).To(Equal(ServiceName1), "instancer's service name should be correct")

		// get same instancer again
		another, e := client.Instancer(ServiceName1)
		g.Expect(e).To(Succeed(), "re-getting instancer should not fail")
		g.Expect(another).To(Equal(instancer), "instancer with same service should be reused")

		// empty service name
		_, e = client.Instancer("")
		g.Expect(e).To(HaveOccurred(), "instancer without service name should fail")

		// try close
		// note: we wait for the instancer finish the intial service update to ensure recorded HTTP order
		_ = instancer.Service()
		e = client.(io.Closer).Close()
		g.Expect(e).To(Succeed(), "closing client should not fail")

		// after close
		_, e = instancer.Instances(nil)
		g.Expect(e).To(BeEquivalentTo(discovery.ErrInstancerStopped), "instancer should fail after client is closed")

	}
}

func SubTestWithDefaultSelector(di *TestDiscoveryDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		client := consulsd.NewDiscoveryClient(ctx, di.Consul, func(opt *consulsd.ClientConfig) {
			opt.DefaultSelector = discovery.InstanceIsHealthy()
			opt.Verbose = true
		})
		defer func() { _ = client.(io.Closer).Close() }()
		instancer, e := client.Instancer(ServiceName1)
		g.Expect(e).To(Succeed(), "getting instancer should not fail")
		g.Expect(instancer).ToNot(BeNil(), "instancer should not be nil")

		// via service
		svc := instancer.Service()
		g.Expect(svc).ToNot(BeNil(), "instancer should return non-nil service")
		g.Expect(svc.Insts).To(HaveLen(2), "instancer should return services with all matching instances")

		// without additional selector
		TryInstancerWithMatcher(g, instancer, nil, []*MockedService{
			&MockedServices[ServiceName1][0], &MockedServices[ServiceName1][1],
		})

		// with additional selector
		TryInstancerWithMatcher(g, instancer, discovery.InstanceWithTag("LEGACY=true", true), []*MockedService{
			&MockedServices[ServiceName1][0],
		})
	}
}

func SubTestWithSelectors(di *TestDiscoveryDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		client := consulsd.NewDiscoveryClient(ctx, di.Consul)
		defer func() { _ = client.(io.Closer).Close() }()
		instancer, e := client.Instancer(ServiceName1)
		g.Expect(e).To(Succeed(), "getting instancer should not fail")
		g.Expect(instancer).ToNot(BeNil(), "instancer should not be nil")

		// via service
		svc := instancer.Service()
		g.Expect(svc).ToNot(BeNil(), "instancer should return non-nil service")
		g.Expect(svc.Insts).To(HaveLen(3), "instancer should return services with all instances")

		// with version
		TryInstancerWithMatcher(g, instancer, discovery.InstanceWithVersion("0.0.1"), []*MockedService{
			&MockedServices[ServiceName1][0],
			&MockedServices[ServiceName1][2],
		})

		// with specific health status
		TryInstancerWithMatcher(g, instancer, discovery.InstanceWithHealth(discovery.HealthCritical), []*MockedService{
			&MockedServices[ServiceName1][2],
		})

		// with tag
		TryInstancerWithMatcher(g, instancer, discovery.InstanceWithTag("legacy=true", false), []*MockedService{
			&MockedServices[ServiceName1][0],
			&MockedServices[ServiceName1][2],
		})

		// with tag KV
		TryInstancerWithMatcher(g, instancer, discovery.InstanceWithTagKV("LeGaCy", "true", true), []*MockedService{
			&MockedServices[ServiceName1][0],
			&MockedServices[ServiceName1][2],
		})

		// with meta
		TryInstancerWithMatcher(g, instancer, discovery.InstanceWithMetaKV("instance", "2"), []*MockedService{
			&MockedServices[ServiceName1][1],
		})

		// with composite matcher
		TryInstancerWithMatcher(g, instancer,
			discovery.InstanceWithTagKV("legacy", "true", false).
				And(discovery.InstanceIsHealthy()).
				Or(discovery.InstanceWithMetaKV("instance", "2")),
			[]*MockedService{
				&MockedServices[ServiceName1][0],
				&MockedServices[ServiceName1][1],
			})

		// with properties
		props := consulsd.SelectorProperties{
			Tags: []string{"LeGaCy=true"},
			Meta: map[string]string{"instance": "3"},
		}
		TryInstancerWithMatcher(g, instancer, consulsd.InstanceWithProperties(&props), []*MockedService{
			&MockedServices[ServiceName1][2],
		})
	}
}

func SubTestWithServiceUpdates(di *TestDiscoveryDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		client := consulsd.NewDiscoveryClient(ctx, di.Consul, func(opt *consulsd.ClientConfig) {
			opt.Verbose = true
		})
		defer func() { _ = client.(io.Closer).Close() }()
		instancer, e := client.Instancer(ServiceName1)
		g.Expect(e).To(Succeed(), "getting instancer should not fail")
		g.Expect(instancer).ToNot(BeNil(), "instancer should not be nil")

		// try some invocation
		TryInstancerWithMatcher(g, instancer, discovery.InstanceIsHealthy(), []*MockedService{
			&MockedServices[ServiceName1][0], &MockedServices[ServiceName1][1],
		})

		// add a callback
		var cbKey = struct{}{}
		var wg sync.WaitGroup
		//wg.Add(1) // registering callback would trigger it once
		instancer.RegisterCallback(cbKey, func(source discovery.Instancer) {
			wg.Done()
		})

		// make some service changes
		wg.Add(1)
		update := MockedServices[ServiceName1][1]
		update.Healthy = false
		e = di.Registrar.Deregister(ctx, NewTestRegistration(&update))
		g.Expect(e).To(Succeed(), "de-registering service should not fail")

		// try again
		wg.Wait()
		TryInstancerWithMatcher(g, instancer, discovery.InstanceIsHealthy(), []*MockedService{
			&MockedServices[ServiceName1][0],
		})
	}
}

func SubTestWithGoKitCompatibility(di *TestDiscoveryDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		client := consulsd.NewDiscoveryClient(ctx, di.Consul, func(opt *consulsd.ClientConfig) {
			opt.Verbose = true
		})
		defer func() { _ = client.(io.Closer).Close() }()
		v, e := client.Instancer(ServiceName1)
		g.Expect(e).To(Succeed(), "getting instancer should not fail")
		g.Expect(v).To(BeAssignableToTypeOf(&consulsd.Instancer{}), "instancer should be correct type")
		instancer := v.(*consulsd.Instancer)

		// register event channel
		eventCh := make(chan sd.Event)
		var lastEvent sd.Event
		var eventLock sync.RWMutex
		defer close(eventCh)
		go func() {
			for evt := range eventCh {
				if !reflect.ValueOf(evt).IsZero() {
					eventLock.Lock()
					lastEvent = evt
					eventLock.Unlock()
				}
			}
		}()
		instancer.Register(eventCh)
		defer instancer.Deregister(eventCh)

		// try some invocation
		TryInstancerWithMatcher(g, instancer, discovery.InstanceIsHealthy(), []*MockedService{
			&MockedServices[ServiceName1][0], &MockedServices[ServiceName1][1],
		})

		// make some service changes
		before := len(lastEvent.Instances)
		update := MockedServices[ServiceName1][1]
		update.Healthy = false
		e = di.Registrar.Deregister(ctx, NewTestRegistration(&update))
		g.Expect(e).To(Succeed(), "de-registering service should not fail")

		// wait for event channel to trigger
		timeoutCtx, cancelFn := context.WithTimeout(ctx, 5*time.Second)
		defer cancelFn()
		for {
			eventLock.RLock()
			updated := len(lastEvent.Instances) != before
			eventLock.RUnlock()
			if updated {
				break
			}
			time.Sleep(50 * time.Millisecond)
			select {
			case <-timeoutCtx.Done():
				t.Errorf("go-kit event is not recieved after service updates")
				return
			default:
			}
		}
		TryInstancerWithMatcher(g, instancer, discovery.InstanceIsHealthy(), []*MockedService{
			&MockedServices[ServiceName1][0],
		})
	}
}

/*************************
	Helpers
 *************************/

type MockedService struct {
	Name    string
	Port    int
	Tags    []string
	Meta    map[string]string
	Healthy bool
}

func OrderedDoForEachMockedService(fn func(reg *consulsd.ServiceRegistration)) {
	regs := make([]*consulsd.ServiceRegistration, 0, 6)
	for _, svcs := range MockedServices {
		for _, svc := range svcs {
			regs = append(regs, NewTestRegistration(&svc))
		}
	}
	sort.SliceStable(regs, func(i, j int) bool {
		return regs[i].Port() < regs[j].Port()
	})
	for _, reg := range regs {
		fn(reg)
	}
}

func NewTestRegistration(svc *MockedService) *consulsd.ServiceRegistration {
	registration := consulsd.NewRegistration(context.Background(), func(cfg *consulsd.RegistrationConfig) {
		cfg.ApplicationName = svc.Name
		cfg.IPAddress = "127.0.0.1"
		cfg.NetworkInterface = "lo"
		cfg.Port = svc.Port
		cfg.Tags = svc.Tags
	}).(*consulsd.ServiceRegistration)
	registration.Check = nil
	if svc.Healthy {
		registration.Check = nil
	} else {
		registration.Check = &api.AgentServiceCheck{
			HTTP:                 "http://localhost:8888/",
			Interval:             "10s",
			SuccessBeforePassing: 1,
		}
	}
	for k, v := range svc.Meta {
		registration.SetMeta(k, v)
	}
	registration.SetID(fmt.Sprintf("%s-%d", svc.Name, svc.Port))
	return registration
}

func TryInstancerWithMatcher(g *gomega.WithT, instancer discovery.Instancer, matcher discovery.InstanceMatcher, expected []*MockedService) {
	insts, e := instancer.Instances(matcher)
	g.Expect(e).To(Succeed(), "Instances should not fail")
	g.Expect(insts).To(HaveLen(len(expected)), "instancer should return correct # of instances")
	for _, svc := range expected {
		expectedID := fmt.Sprintf("%s-%d", svc.Name, svc.Port)
		var found bool
		for _, inst := range insts {
			if inst.ID != expectedID {
				continue
			}
			g.Expect(inst.Service).To(Equal(svc.Name), "instance with ID [%s] should have correct %s", expectedID, "Service")
			g.Expect(inst.Address).To(Equal("127.0.0.1"), "instance with ID [%s] should have correct %s", expectedID, "Address")
			g.Expect(inst.Port).To(Equal(svc.Port), "instance with ID [%s] should have correct %s", expectedID, "Port")
			found = true
		}
		g.Expect(found).To(BeTrue(), "instance with ID [%s] should exists", expectedID)
	}
}
