package actuator

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
)

const (
	OperationRead OperationMode = iota
	OperationWrite
)
type OperationMode int

// OperationFunc is a func that have following signature:
// 	func(ctx context.Context, input StructsOrPointerType1) (StructsOrPointerType2, error)
// where
//	- StructsOrPointerType1 and StructsOrPointerType2 can be any structs or struct pointers
//  - input might be ignored by particular Endpoint impl.
//  - 1st output is optional for "write" operations
//
// Note: golang doesn't have generics yet...
type OperationFunc interface{}

type Operation interface {
	Mode() OperationMode
	Func() OperationFunc
	Matches(ctx context.Context, mode OperationMode, input interface{}) bool
	Execute(ctx context.Context, input interface{}) (interface{}, error)
}

type Endpoint interface {
	Id() string
	EnabledByDefault() bool
	Operations() []Operation
}

type WebEndpoint interface {
	Mappings(op Operation, group string) ([]web.Mapping, error)
}

type EndpointExecutor interface {
	ReadOperation(ctx context.Context, input interface{}) (interface{}, error)
	WriteOperation(ctx context.Context, input interface{}) (interface{}, error)
}

type WebEndpoints map[string][]web.Mapping

func (w WebEndpoints) EndpointIDs() (ret []string) {
	ret = make([]string, 0, len(w))
	for k, _ := range w {
		ret = append(ret, k)
	}
	return
}

// Paths returns all path patterns of given endpoint ID.
// only web.RoutedMapping & web.StaticMapping is possible to extract this information
func (w WebEndpoints) Paths(id string) []string {
	mappings, ok := w[id]
	if !ok {
		return []string{}
	}

	paths := utils.NewStringSet()
	for _, v := range mappings {
		switch v.(type) {
		case web.RoutedMapping:
			paths.Add(v.(web.RoutedMapping).Path())
		case web.StaticMapping:
			paths.Add(v.(web.StaticMapping).Path())
		}
	}
	return paths.Values()
}
