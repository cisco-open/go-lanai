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
	AppCtx *bootstrap.ApplicationContext
	Client *ClientMock
}

type SDMockOptions func(opt *SDMockOption)
type SDMockOption struct {
	FS               fs.FS
	DefPath          string
	PropertiesPrefix string
}

func WithMockedSD(opts ...SDMockOptions) test.Options {
	var di DI
	testOpts := []test.Options{
		apptest.WithFxOptions(
			fx.Provide(ProvideDiscoveryClient),
		),
		apptest.WithDI(&di),
	}
	var opt SDMockOption
	for _, fn := range opts {
		fn(&opt)
	}

	// load service definitions
	switch {
	case opt.FS != nil && opt.DefPath != "":
		testOpts = append(testOpts, test.SubTestSetup(SetupServicesWithFile(&di, opt.FS, opt.DefPath)))
	default:
		testOpts = append(testOpts, test.SubTestSetup(SetupServicesWithProperties(&di, opt.PropertiesPrefix)))
	}
	return test.WithOptions(testOpts...)
}

// LoadDefinition load service discovery mocking from file system, this override DefinitionWithPrefix
func LoadDefinition(fsys fs.FS, path string) SDMockOptions {
	return func(opt *SDMockOption) {
		opt.FS = fsys
		opt.DefPath = path
	}
}

// DefinitionWithPrefix load service discovery mocking from application properties, with given prefix
func DefinitionWithPrefix(prefix string) SDMockOptions {
	return func(opt *SDMockOption) {
		opt.PropertiesPrefix = prefix
	}
}

func ProvideDiscoveryClient(ctx *bootstrap.ApplicationContext) (discovery.Client, *ClientMock) {
	c := NewMockClient(ctx)
	return c, c
}

// SetupServicesWithFile is a test setup function that read service definitions from a YAML file and mock the discovery client
func SetupServicesWithFile(di *DI, fsys fs.FS, path string) test.SetupFunc {
	return func(ctx context.Context, t *testing.T) (context.Context, error) {
		if di == nil || di.Client == nil {
			return nil, errors.New("discovery client mock is not available")
		}
		e := MockServicesFromFile(di.Client, fsys, path)
		return ctx, e
	}
}

// SetupServicesWithProperties is a test setup function that read service definitions from properties and mock the discovery client
func SetupServicesWithProperties(di *DI, prefix string) test.SetupFunc {
	return func(ctx context.Context, t *testing.T) (context.Context, error) {
		if di == nil || di.Client == nil {
			return nil, errors.New("discovery client mock is not available")
		}
		e := MockServicesFromProperties(di.Client, di.AppCtx.Config(), prefix)
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

// MockServicesFromProperties bind mocked service definitions from properties with given prefix and mock ClientMock
func MockServicesFromProperties(client *ClientMock, appCfg bootstrap.ApplicationConfig, prefix string) error {
	var services map[string][]*discovery.Instance
	if e := appCfg.Bind(&services, prefix); e != nil {
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
