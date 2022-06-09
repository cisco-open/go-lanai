// Package sdtest
// test utilities to mock service discovery client
package sdtest

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/discovery"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/apptest"
	"errors"
	"github.com/ghodss/yaml"
	"github.com/imdario/mergo"
	"go.uber.org/fx"
	"io"
	"io/fs"
	"testing"
)

type DI struct {
	fx.In
	Client *ClientMock
}

type SDMockOptions func(opt *SDMockOption)
type SDMockOption struct {
	FS      fs.FS
	DefPath string
}

func WithMockedSD(opts ...SDMockOptions) test.Options {
	testOpts := []test.Options{
		apptest.WithFxOptions(
			fx.Provide(ProvideDiscoveryClient),
		),
	}
	var opt SDMockOption
	for _, fn := range opts {
		fn(&opt)
	}
	if opt.FS != nil && opt.DefPath != "" {
		var di DI
		testOpts = append(testOpts,
			apptest.WithDI(&di),
			test.SubTestSetup(SetupServices(&di, opt.FS, opt.DefPath)),
		)
	}
	return test.WithOptions(testOpts...)
}

func LoadDefinition(fsys fs.FS, path string) SDMockOptions {
	return func(opt *SDMockOption) {
		opt.FS = fsys
		opt.DefPath = path
	}
}

func ProvideDiscoveryClient(ctx *bootstrap.ApplicationContext) (discovery.Client, *ClientMock) {
	c := NewMockClient(ctx)
	return c, c
}

// SetupServices is a test setup function that read service definitions and mock the discovery client
func SetupServices(di *DI, fsys fs.FS, path string) test.SetupFunc {
	return func(ctx context.Context, t *testing.T) (context.Context, error) {
		if di == nil || di.Client == nil {
			return nil, errors.New("discovery client mock is not available")
		}
		e := MockServicesFromFile(di.Client, fsys, path)
		return ctx, e
	}
}

// MockServicesFromFile read YAML file for mocked service definition and mock ClientMock
func MockServicesFromFile(client *ClientMock, fsys fs.FS, path string) error {
	var services map[string][]*discovery.Instance
	file, e := fsys.Open(path)
	if e != nil {
		return e
	}
	defer func() { _ = file.Close() }()
	data, e := io.ReadAll(file)
	if e != nil {
		return e
	}
	if e := yaml.Unmarshal(data, &services); e != nil {
		return e
	}
	return MockServices(client, services)
}

// MockServices mocks given ClientMock with given services. The key is the service name
func MockServices(client *ClientMock, services map[string][]*discovery.Instance) (err error) {
	for k, insts := range services {
		var i int
		client.MockService(k, len(insts), func(inst *discovery.Instance) {
			defer func() { i++ }()
			def := insts[i]
			if e := mergo.Merge(inst, def, mergo.WithAppendSlice, mergo.WithSliceDeepCopy); e != nil {
				err = e
			}
			if def.Health == discovery.HealthAny {
				inst.Health = discovery.HealthPassing
			}
		})
		if err != nil {
			break
		}
	}
	return
}
