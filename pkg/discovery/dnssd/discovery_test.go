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

package dnssd_test

import (
	"context"
	"github.com/cisco-open/go-lanai/pkg/bootstrap"
	"github.com/cisco-open/go-lanai/pkg/discovery"
	"github.com/cisco-open/go-lanai/pkg/discovery/dnssd"
	"github.com/cisco-open/go-lanai/test"
	"github.com/cisco-open/go-lanai/test/apptest"
	"github.com/go-kit/kit/sd"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
	"io"
	"net"
	"reflect"
	"regexp"
	"sort"
	"strconv"
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
}

/*************************
	Tests
 *************************/

type TestDiscoveryDI struct {
	fx.In
	AppContext *bootstrap.ApplicationContext
}

func TestDiscoveryClient(t *testing.T) {
	di := TestDiscoveryDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		WithDNSServer(),
		apptest.WithDI(&di),
		test.SubTestSetup(SetupTestServices()),
		test.SubTestTeardown(TeardownTestServices()),
		test.GomegaSubTest(SubTestInstancerManagement(&di), "TestInstancerManagement"),
		test.GomegaSubTest(SubTestWithoutProto(&di), "TestWithoutProto"),
		test.GomegaSubTest(SubTestWithProtoAndService(&di), "TestWithProtoAndService"),
		test.GomegaSubTest(SubTestWithFallback(&di), "TestWithFallback"),
		test.GomegaSubTest(SubTestWithServiceUpdates(&di), "TestWithServiceUpdates"),
		test.GomegaSubTest(SubTestWithGoKitCompatibility(&di), "TestWithGoKitCompatibility"),
	)
}

/*************************
	Sub-Test Cases
 *************************/

func SetupTestServices() test.SetupFunc {
	return func(ctx context.Context, t *testing.T) (context.Context, error) {
		OrderedDoForEachMockedService(func(reg *MockedSRV) {
			CurrentMockedDNSServer(ctx).RegisterSRV(reg)
		})
		return ctx, nil
	}
}

func TeardownTestServices() test.TeardownFunc {
	return func(ctx context.Context, t *testing.T) error {
		OrderedDoForEachMockedService(func(reg *MockedSRV) {
			CurrentMockedDNSServer(ctx).DeregisterSRV(reg)
		})
		return nil
	}
}

func SubTestInstancerManagement(_ *TestDiscoveryDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		client := dnssd.NewDiscoveryClient(ctx, func(opt *dnssd.ClientConfig) {
			opt.DNSServerAddr = CurrentMockedDNSAddr(ctx)
			opt.FQDNTemplate = "{{.ServiceName}}.test.mock"
			opt.Verbose = true
		})
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
		_ = instancer.Service()
		e = client.(io.Closer).Close()
		g.Expect(e).To(Succeed(), "closing client should not fail")

		// after close
		_, e = instancer.Instances(nil)
		g.Expect(e).To(BeEquivalentTo(discovery.ErrInstancerStopped), "instancer should fail after client is closed")
	}
}

func SubTestWithoutProto(_ *TestDiscoveryDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		client := dnssd.NewDiscoveryClient(ctx, func(opt *dnssd.ClientConfig) {
			opt.DNSServerAddr = CurrentMockedDNSAddr(ctx)
			opt.FQDNTemplate = "{{.ServiceName}}.test.mock"
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

		//with additional selector
		TryInstancerWithMatcher(g, instancer, discovery.InstanceWithHealth(discovery.HealthMaintenance), []*MockedService{})
	}
}

func SubTestWithProtoAndService(_ *TestDiscoveryDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		client := dnssd.NewDiscoveryClient(ctx, func(opt *dnssd.ClientConfig) {
			opt.DNSServerAddr = CurrentMockedDNSAddr(ctx)
			opt.FQDNTemplate = "{{.ServiceName}}.test.mock"
			opt.SRVProto = TestProto
			opt.SRVService = TestService
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

		//with additional selector
		TryInstancerWithMatcher(g, instancer, discovery.InstanceWithHealth(discovery.HealthCritical), []*MockedService{})
	}
}

func SubTestWithFallback(_ *TestDiscoveryDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		client := dnssd.NewDiscoveryClient(ctx, func(opt *dnssd.ClientConfig) {
			opt.DNSServerAddr = CurrentMockedDNSAddr(ctx)
			opt.FQDNTemplate = "{{.ServiceName}}.test.mock"
			opt.SRVProto = TestProto
			opt.SRVService = TestService
			opt.Verbose = true
			opt.FallbackHostMappings = []dnssd.HostMapping{
				{
					ServiceRegex: regexp.MustCompilePOSIX(ServiceName2),
					Hosts:        []string{"inst-1.{{.ServiceName}}.test.mock:9999", "inst-2.{{.ServiceName}}.test.mock:8888"},
				},
				{
					ServiceRegex: regexp.MustCompilePOSIX(`.+`),
					Hosts:        []string{"{{.ServiceName}}.test.mock"},
				},
			}
		})
		defer func() { _ = client.(io.Closer).Close() }()
		// Matched Service
		instancer, e := client.Instancer(ServiceName2)
		g.Expect(e).To(Succeed(), "getting instancer should not fail")
		g.Expect(instancer).ToNot(BeNil(), "instancer should not be nil")

		// via service
		svc := instancer.Service()
		g.Expect(svc).ToNot(BeNil(), "instancer should return non-nil service")
		g.Expect(svc.Insts).To(HaveLen(2), "instancer should return services with fallback instances")

		// without additional selector
		TryInstancerWithMatcher(g, instancer, nil, []*MockedService{
			{
				AlternativeID: "inst-1." + ServiceName2 + ".test.mock:9999",
				AlternativeAddr: "inst-1." + ServiceName2 + ".test.mock",
				Name:          ServiceName2,
				Port:          9999,
				Healthy:       true,
			},
			{
				AlternativeID: "inst-2." + ServiceName2 + ".test.mock:8888",
				AlternativeAddr: "inst-2." + ServiceName2 + ".test.mock",
				Name:          ServiceName2,
				Port:          8888,
				Healthy:       true,
			},
		})
		//with additional selector
		TryInstancerWithMatcher(g, instancer, discovery.InstanceWithHealth(discovery.HealthCritical), []*MockedService{})

		// Default Service
		instancer, e = client.Instancer("unknown-service")
		g.Expect(e).To(Succeed(), "getting instancer should not fail")
		g.Expect(instancer).ToNot(BeNil(), "instancer should not be nil")

		// via service
		svc = instancer.Service()
		g.Expect(svc).ToNot(BeNil(), "instancer should return non-nil service")
		g.Expect(svc.Insts).To(HaveLen(1), "instancer should return services with fallback instances")

		// without additional selector
		TryInstancerWithMatcher(g, instancer, nil, []*MockedService{
			{
				AlternativeID: "unknown-service.test.mock:0",
				AlternativeAddr: "unknown-service.test.mock",
				Name:          "unknown-service",
				Port:          0,
				Healthy:       true,
			},
		})
		//with additional selector
		TryInstancerWithMatcher(g, instancer, discovery.InstanceWithHealth(discovery.HealthCritical), []*MockedService{})
	}
}

func SubTestWithServiceUpdates(_ *TestDiscoveryDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		client := dnssd.NewDiscoveryClient(ctx, func(opt *dnssd.ClientConfig) {
			opt.DNSServerAddr = CurrentMockedDNSAddr(ctx)
			opt.FQDNTemplate = "{{.ServiceName}}.test.mock"
			opt.Verbose = true
			opt.RefreshInterval = 50 * time.Millisecond
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
		instancer.RegisterCallback(cbKey, func(source discovery.Instancer) {
			wg.Done()
		})

		// make some service changes
		wg.Add(1)
		update := MockedServices[ServiceName1][1]
		update.Healthy = false
		CurrentMockedDNSServer(ctx).DeregisterSRV(NewMockedSRV(&update))

		// try again
		wg.Wait()
		TryInstancerWithMatcher(g, instancer, discovery.InstanceIsHealthy(), []*MockedService{
			&MockedServices[ServiceName1][0],
		})
	}
}

func SubTestWithGoKitCompatibility(_ *TestDiscoveryDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		client := dnssd.NewDiscoveryClient(ctx, func(opt *dnssd.ClientConfig) {
			opt.DNSServerAddr = CurrentMockedDNSAddr(ctx)
			opt.FQDNTemplate = "{{.ServiceName}}.test.mock"
			opt.Verbose = true
			opt.RefreshInterval = 50 * time.Millisecond
		})
		defer func() { _ = client.(io.Closer).Close() }()
		v, e := client.Instancer(ServiceName1)
		g.Expect(e).To(Succeed(), "getting instancer should not fail")
		g.Expect(v).To(BeAssignableToTypeOf(&dnssd.Instancer{}), "instancer should be correct type")
		instancer := v.(*dnssd.Instancer)

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
		CurrentMockedDNSServer(ctx).DeregisterSRV(NewMockedSRV(&update))

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
	AlternativeID   string // only used for assertion
	AlternativeAddr string // only used for assertion
	Name            string
	Port            int
	Tags            []string
	Meta            map[string]string
	Healthy         bool
}

func OrderedDoForEachMockedService(fn func(reg *MockedSRV)) {
	regs := make([]*MockedSRV, 0, 6)
	for _, svcs := range MockedServices {
		for _, svc := range svcs {
			regs = append(regs, NewMockedSRV(&svc))
		}
	}
	sort.SliceStable(regs, func(i, j int) bool {
		return regs[i].Port < regs[j].Port
	})
	for _, reg := range regs {
		fn(reg)
	}
}

func TryInstancerWithMatcher(g *gomega.WithT, instancer discovery.Instancer, matcher discovery.InstanceMatcher, expected []*MockedService) {
	insts, e := instancer.Instances(matcher)
	g.Expect(e).To(Succeed(), "Instances should not fail")
	g.Expect(insts).To(HaveLen(len(expected)), "instancer should return correct # of instances")
	for _, svc := range expected {
		expectedID := svc.AlternativeID
		if len(expectedID) == 0 {
			expectedID = net.JoinHostPort(NewMockedSRV(svc).Address(), strconv.Itoa(svc.Port))
		}
		var found bool
		for _, inst := range insts {
			if inst.ID != expectedID {
				continue
			}
			expectedAddr := svc.AlternativeAddr
			if len(expectedAddr) == 0 {
				expectedAddr = AddrToDomain("127.0.0.1", ServiceFQDN(svc.Name)+".")
			}
			g.Expect(inst.Service).To(Equal(svc.Name), "instance with ID [%s] should have correct %s", expectedID, "Service")
			g.Expect(inst.Address).To(Equal(expectedAddr), "instance with ID [%s] should have correct %s", expectedID, "Address")
			g.Expect(inst.Port).To(Equal(svc.Port), "instance with ID [%s] should have correct %s", expectedID, "Port")
			found = true
		}
		g.Expect(found).To(BeTrue(), "instance with ID [%s] should exists", expectedID)
	}
}
