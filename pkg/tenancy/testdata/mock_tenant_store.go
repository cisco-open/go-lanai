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

package testdata

import (
	"context"
	th_loader "cto-github.cisco.com/NFV-BU/go-lanai/pkg/tenancy/loader"
	"fmt"
	"github.com/ghodss/yaml"
	"github.com/google/uuid"
	"io"
	"io/fs"
)

type TestData struct {
	Tenants []TestTenant         `json:"tenants"`
	UUIDs   map[string]uuid.UUID `json:"uuids"`
}

type TestTenant struct {
	Name        string `json:"name"`
	Parent      string `json:"parent"`
	uuidMapping map[string]uuid.UUID
}

func (t TestTenant) GetId() string {
	if len(t.Name) == 0 {
		return ""
	}
	return t.uuidMapping[t.Name].String()
}

func (t TestTenant) GetParentId() string {
	if len(t.Parent) == 0 {
		return ""
	}
	return t.uuidMapping[t.Parent].String()
}

type TestTenantStore struct {
	TestData
	SourceFS   fs.FS
	SourcePath string
}

func (s *TestTenantStore) Reset(srcFS fs.FS, srcPath string) {
	s.Tenants = nil
	s.UUIDs = nil
	s.SourceFS = srcFS
	s.SourcePath = srcPath
}

func (s *TestTenantStore) IDof(tenant string) string {
	if len(tenant) == 0 || s.UUIDs == nil {
		return ""
	}
	return s.UUIDs[tenant].String()
}

func (s *TestTenantStore) GetIterator(_ context.Context) (th_loader.TenantIterator, error) {
	if len(s.SourcePath) == 0 || s.SourceFS == nil {
		return &TestTenantIterator{Tenants: []TestTenant{}}, nil
	}

	if len(s.Tenants) == 0 {
		data, e := fs.ReadFile(s.SourceFS, s.SourcePath)
		if e != nil {
			return nil, fmt.Errorf("unable to load test tenants file: %v", e)
		}
		if e := yaml.Unmarshal(data, &s.TestData); e != nil {
			return nil, fmt.Errorf("unable to parse test tenants file: %v", e)
		}
		for i := range s.Tenants {
			s.Tenants[i].uuidMapping = s.UUIDs
		}
	}
	return &TestTenantIterator{Tenants: s.Tenants}, nil
}

type TestTenantIterator struct {
	Tenants []TestTenant
}

func (i *TestTenantIterator) Next() bool {
	return len(i.Tenants) != 0
}

func (i *TestTenantIterator) Scan(_ context.Context) (th_loader.Tenant, error) {
	if len(i.Tenants) == 0 {
		return nil, io.EOF
	}
	defer func() {
		i.Tenants = i.Tenants[1:]
	}()
	return i.Tenants[0], nil
}

func (i *TestTenantIterator) Close() error {
	return nil
}

func (i *TestTenantIterator) Err() error {
	return nil
}
