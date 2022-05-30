package discoverymock

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/discovery"
	"time"
)

type ClientMock struct {
	ctx context.Context
	Instancers map[string]*InstancerMock
}

func NewMockClient(ctx context.Context) *ClientMock {
	return &ClientMock{
		ctx: ctx,
		Instancers: map[string]*InstancerMock{},
	}
}

/* discovery.Client implementation */

func (c *ClientMock) Context() context.Context {
	return c.ctx
}

func (c *ClientMock) Instancer(serviceName string) (discovery.Instancer, error) {
	if instancer, ok := c.Instancers[serviceName]; ok {
		return instancer, nil
	}

	instancer := NewMockInstancer(c.ctx, serviceName)
	c.Instancers[serviceName] = instancer
	return instancer, nil
}

/* Addtional mock methods */

func (c *ClientMock) MockService(svcName string, count int, opts ...InstanceMockOptions) []*discovery.Instance {
	instancer, _ := c.Instancer(svcName)
	return instancer.(*InstancerMock).MockInstances(count, opts...)
}

func (c *ClientMock) UpdateMockedService(svcName string, matcher InstanceMockMatcher, opts ...InstanceMockOptions) (count int) {
	instancer, ok := c.Instancers[svcName]
	if !ok {
		return 0
	}
	return instancer.UpdateInstances(matcher, opts...)
}

func (c *ClientMock) MockError(svcName string, what error, when time.Time) {
	instancer, _ := c.Instancer(svcName)
	instancer.(*InstancerMock).MockError(what, when)
}