package mocks

import (
	"container/list"
	"context"
	"errors"
	"github.com/google/uuid"
)

type TenancyRelation struct {
	Child uuid.UUID
	Parent uuid.UUID
}

type MockTenancyAccessor struct {
	ParentLookup      map[string]string
	ChildrenLookup    map[string][]string
	DescendantsLookup map[string][]string
	AncestorsLookup   map[string][]string
	Root 			  string
}

func NewMockTenancyAccessor (tenantRelations []TenancyRelation, root uuid.UUID) *MockTenancyAccessor {
	m := &MockTenancyAccessor{
	}
	m.Reset(tenantRelations, root)
	return m
}

func (m *MockTenancyAccessor) Reset (tenantRelations []TenancyRelation, root uuid.UUID) {
	m.ParentLookup = make(map[string]string)
	m.ChildrenLookup = make(map[string][]string)
	m.DescendantsLookup = make(map[string][]string)
	m.AncestorsLookup = make(map[string][]string)
	m.Root = root.String()


	//build the parent and children lookup
	for _, r := range tenantRelations {
		m.ParentLookup[r.Child.String()] = r.Parent.String()

		children := m.ChildrenLookup[r.Parent.String()]
		children = append(children, r.Child.String())
		m.ChildrenLookup[r.Parent.String()] = children
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
	return true
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
