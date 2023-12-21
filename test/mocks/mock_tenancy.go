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

package mocks

import (
	"container/list"
	"context"
	"errors"
	"github.com/google/uuid"
)

// TenancyRelation
// Deprecated: use the string version instead
type TenancyRelation struct {
	Child  uuid.UUID
	Parent uuid.UUID
}

type TenancyRelationWithStrId struct {
	ChildId  string
	ParentId string
}

type MockTenancyAccessor struct {
	ParentLookup      map[string]string
	ChildrenLookup    map[string][]string
	DescendantsLookup map[string][]string
	AncestorsLookup   map[string][]string
	Root              string
	Isloaded          bool
}

// NewMockTenancyAccessor
// Deprecated: Use string version instead
func NewMockTenancyAccessor(tenantRelations []TenancyRelation, root uuid.UUID) *MockTenancyAccessor {
	m := &MockTenancyAccessor{}
	// default
	m.Isloaded = true
	m.Reset(tenantRelations, root)
	return m
}

func NewMockTenancyAccessorUsingStrIds(tenantRelations []TenancyRelationWithStrId, root string) *MockTenancyAccessor {
	m := &MockTenancyAccessor{}
	m.Isloaded = true
	m.ResetWithStrIds(tenantRelations, root)
	return m
}

// Reset
// Deprecated: Use the str version instead
func (m *MockTenancyAccessor) Reset(tenantRelations []TenancyRelation, root uuid.UUID) {
	var trWithStrId []TenancyRelationWithStrId
	for _, tr := range tenantRelations {
		trWithStrId = append(trWithStrId, TenancyRelationWithStrId{
			ChildId:  tr.Child.String(),
			ParentId: tr.Parent.String(),
		})
	}
	rootStrId := root.String()
	m.ResetWithStrIds(trWithStrId, rootStrId)
}

func (m *MockTenancyAccessor) ResetWithStrIds(tenantRelations []TenancyRelationWithStrId, root string) {
	m.ParentLookup = make(map[string]string)
	m.ChildrenLookup = make(map[string][]string)
	m.DescendantsLookup = make(map[string][]string)
	m.AncestorsLookup = make(map[string][]string)
	m.Root = root

	//build the parent and children lookup
	for _, r := range tenantRelations {
		m.ParentLookup[r.ChildId] = r.ParentId

		children := m.ChildrenLookup[r.ParentId]
		children = append(children, r.ChildId)
		m.ChildrenLookup[r.ParentId] = children
	}

	//build the ancestor lookup
	for child, _ := range m.ParentLookup {
		var ancestors []string
		tenantId := child
		for {
			parent, ok := m.ParentLookup[tenantId]
			if ok {
				ancestors = append(ancestors, parent)
				tenantId = parent
			} else {
				break
			}
		}
		m.AncestorsLookup[child] = ancestors
	}

	//build the descendant lookup
	for parent, _ := range m.ChildrenLookup {
		var descendants []string

		idsToVisit := list.New()
		idsToVisit.PushBack(parent)

		for idsToVisit.Len() != 0 {
			id := idsToVisit.Front()
			idsToVisit.Remove(id)
			if children, ok := m.ChildrenLookup[id.Value.(string)]; ok {
				for _, c := range children {
					idsToVisit.PushBack(c)
				}
				descendants = append(descendants, children...)
			}
		}
		m.DescendantsLookup[parent] = descendants
	}
}

func (m *MockTenancyAccessor) GetParent(ctx context.Context, tenantId string) (string, error) {
	if parent, ok := m.ParentLookup[tenantId]; ok {
		return parent, nil
	} else {
		return "", errors.New("parent not found")
	}
}

func (m *MockTenancyAccessor) GetChildren(ctx context.Context, tenantId string) ([]string, error) {
	if children, ok := m.ChildrenLookup[tenantId]; ok {
		return children, nil
	} else {
		return nil, errors.New("children not found")
	}
}

func (m *MockTenancyAccessor) GetAncestors(ctx context.Context, tenantId string) ([]string, error) {
	if tenantId == m.Root {
		return make([]string, 0), nil
	}
	if ancestors, ok := m.AncestorsLookup[tenantId]; ok {
		return ancestors, nil
	} else {
		return nil, errors.New("ancestors not found")
	}
}

func (m *MockTenancyAccessor) GetDescendants(ctx context.Context, tenantId string) ([]string, error) {
	if descendants, ok := m.DescendantsLookup[tenantId]; ok {
		return descendants, nil
	} else {
		return nil, errors.New("descendants not found")
	}
}

func (m *MockTenancyAccessor) GetRoot(ctx context.Context) (string, error) {
	if m.Root != "" {
		return m.Root, nil
	} else {
		return "", errors.New("root not set")
	}
}

func (m *MockTenancyAccessor) IsLoaded(ctx context.Context) bool {
	return m.Isloaded
}

func (m *MockTenancyAccessor) GetTenancyPath(ctx context.Context, tenantId string) ([]uuid.UUID, error) {
	current, err := uuid.Parse(tenantId)
	if err != nil {
		return nil, err
	}
	path := []uuid.UUID{current}

	ancestors, err := m.GetAncestors(ctx, tenantId)
	if err != nil {
		return nil, err
	}

	for _, str := range ancestors {
		id, err := uuid.Parse(str)
		if err != nil {
			return nil, err
		}
		path = append(path, id)
	}

	//reverse the order to that the result is root tenant id -> current tenant id
	//fi is index going forward starting from 0,
	//ri is index going backward starting from last element
	//swap the element at ri and ri
	for fi, ri := 0, len(path)-1; fi < ri; fi, ri = fi+1, ri-1 {
		path[fi], path[ri] = path[ri], path[fi]
	}
	return path, nil
}
