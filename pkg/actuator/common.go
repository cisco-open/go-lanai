package actuator

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/matcher"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/rest"
	"fmt"
	"net/http"
	"reflect"
	"strings"
)

var (
	ctxType = reflect.TypeOf((*context.Context)(nil)).Elem()
	errorType = reflect.TypeOf((*error)(nil)).Elem()
)

/*******************************
	Operation
********************************/
// operation implements Operation, and hold some metadata with reflection
type operation struct {
	mode     OperationMode
	f        OperationFunc
	matcher  matcher.Matcher
	function reflect.Value
	input    reflect.Type
	output   reflect.Type
}

func newOperation(mode OperationMode, opFunc OperationFunc, inputMatchers...matcher.Matcher) *operation {
	var m matcher.Matcher
	switch len(inputMatchers) {
	case 0:
		// do nothing
	case 1:
		m = inputMatchers[0]
	default:
		m = matcher.Or(inputMatchers[0], inputMatchers[1:]...)
	}
	op := operation{
		mode:    mode,
		f:       opFunc,
		matcher: m,
	}
	if e := populateOperationMetadata(opFunc, &op); e != nil {
		panic(e)
	}
	return &op
}

func NewReadOperation(opFunc OperationFunc, inputMatchers...matcher.Matcher) Operation {
	return newOperation(OperationRead, opFunc, inputMatchers...)
}

func NewWriteOperation(opFunc OperationFunc, inputMatchers...matcher.Matcher) Operation {
	return newOperation(OperationWrite, opFunc, inputMatchers...)
}

func (op operation) Mode() OperationMode {
	return op.mode
}

func (op operation) Func() OperationFunc {
	return op.f
}

func (op operation) Matches(ctx context.Context, mode OperationMode, input interface{}) bool {
	if mode != op.mode || !reflect.TypeOf(input).ConvertibleTo(op.input) {
		return false
	}
	if op.matcher == nil {
		return true
	}
	m, e := op.matcher.MatchesWithContext(ctx, input)
	return e != nil || !m
}

func (op operation) Execute(ctx context.Context, input interface{}) (interface{}, error) {
	in := reflect.ValueOf(input).Convert(op.input)
	ret := op.function.Call([]reflect.Value{reflect.ValueOf(ctx), in})
	switch len(ret) {
	case 1:
		return nil, ret[0].Interface().(error)
	case 2:
		return ret[0].Interface(), ret[0].Interface().(error)
	default:
		// find error param
		for _, v := range ret {
			if e, ok := v.Interface().(error); ok {
				return nil, e
			}
		}
		return nil, fmt.Errorf("operation failed with unknown error")
	}
}

func populateOperationMetadata(opFunc OperationFunc, op *operation) error {
	op.function = reflect.ValueOf(opFunc)
	if e := validateFunc(op.function); e != nil {
		return e
	}

	t := op.function.Type()
	op.input = t.In(t.NumIn() - 1)
	if t.NumOut() > 1 {
		op.output = t.Out(0)
	}
	return nil
}

func validateFunc(f reflect.Value) error {
	// since golang doesn't have generics, we have to check the signature at run-time
	if f.Kind() != reflect.Func {
		return fmt.Errorf("OperationFunc must be a function but got %T", f.Interface())
		//return fmt.Errorf("OperationFunc must have signigure 'func(ctx context.Context, input Type1) (Type2, error)', but got %v", f.Interface())
	}

	t := f.Type()
	switch {
	// input params validation
	case t.NumIn() != 2:
		fallthrough
	case !t.In(0).Implements(ctxType):
		fallthrough
	case !isStructOrPtrToStruct(t.In(1)):
		fallthrough
	// Out params validation
	case t.NumOut() < 1 || t.NumOut() > 2:
		fallthrough
	case !t.Out(t.NumOut() - 1).Implements(errorType):
		fallthrough
	case t.NumOut() == 2 && !isSupportedOutputType(t.Out(0)):
		return invalidOpFuncSignatureError(f.Interface())
	}
	return nil
}

func isStructOrPtrToStruct(t reflect.Type) (ret bool) {
	ret = t.Kind() == reflect.Struct
	ret = ret || t.Kind() == reflect.Ptr && t.Elem().Kind() == reflect.Struct
	return
}

func isSupportedOutputType(t reflect.Type) bool {
	if isStructOrPtrToStruct(t) {
		return true
	}
	switch t.Kind() {
	case reflect.Interface:
		fallthrough
	case reflect.Map:
		return true
	}
	return false
}

func invalidOpFuncSignatureError(f interface{}) error {
	return fmt.Errorf("OperationFunc must have signigure 'func(ctx context.Context, input Type1) (Type2, error)', but got %v", f)
}

/*******************************
	EndpointBase
********************************/

// EndpointBase implements EndpointExecutor and partially Endpoint,  and can be embedded into any Endpoint implementation
// it calls Operations using reflect
type EndpointBase struct {
	id               string
	operations       []Operation
	properties       *EndpointsProperties
	enabledByDefault bool
}

type EndpointOptions func(opt *EndpointOption)

type EndpointOption struct {
	Id               string
	Ops              []Operation
	Properties       *EndpointsProperties
	EnabledByDefault bool
}

func MakeEndpointBase(opts...EndpointOptions) EndpointBase {
	opt := EndpointOption{}
	for _, f := range opts {
		f(&opt)
	}
	return EndpointBase{
		id:               opt.Id,
		operations:       opt.Ops,
		properties:       opt.Properties,
		enabledByDefault: opt.EnabledByDefault,
	}
}

func (b EndpointBase) Id() string {
	return b.id
}

func (b EndpointBase) EnabledByDefault() bool {
	return b.enabledByDefault
}

func (b EndpointBase) Operations() []Operation {
	return b.operations
}

func (b EndpointBase) ReadOperation(ctx context.Context, input interface{}) (interface{}, error) {
	for _, op := range b.operations {
		if op.Matches(ctx, OperationRead, input) {
			return op.Execute(ctx, input)
		}
	}
	return nil, fmt.Errorf("unsupported read operation [%s] with input [%v]", b.Id(), input)
}

func (b EndpointBase) WriteOperation(ctx context.Context, input interface{}) (interface{}, error) {
	for _, op := range b.operations {
		if op.Matches(ctx, OperationWrite, input) {
			return op.Execute(ctx, input)
		}
	}
	return nil, fmt.Errorf("unsupported write operation [%s] with input [%v]", b.Id(), input)
}

/*******************************
	WebEndpointBase
********************************/

type MappingPathFunc func(op Operation, props *WebEndpointsProperties) string
type MappingNameFunc func(op Operation) string

// WebEndpointBase is similar to EndpointBase and implements default WebEndpoint mapping
type WebEndpointBase struct {
	EndpointBase
	properties *WebEndpointsProperties
}

func MakeWebEndpointBase(opts...EndpointOptions) WebEndpointBase {
	base := MakeEndpointBase(opts...)
	return WebEndpointBase{
		EndpointBase: base,
		properties: &base.properties.Web,
	}
}

// Mappings implements WebEndpoint
func (b WebEndpointBase) Mappings(op Operation, group string) ([]web.Mapping, error) {
	builder, e := b.RestMappingBuilder(op, group, b.MappingPath, b.MappingName)
	if e != nil {
		return nil, e
	}
	return []web.Mapping{builder.Build()}, nil
}

func (b WebEndpointBase) MappingPath(_ Operation, props *WebEndpointsProperties) string {
	base := strings.Trim(props.BasePath, "/")
	path, ok := props.Mappings[b.id]
	if !ok {
		path = strings.ToLower(b.id)
	}
	path = strings.Trim(path, "/")
	return fmt.Sprintf("/%s/%s", base, path)
}

func (b WebEndpointBase) MappingName(op Operation) string {
	switch op.Mode() {
	case OperationRead:
		return fmt.Sprintf("%s GET", strings.ToLower(b.id))
	case OperationWrite:
		return fmt.Sprintf("%s POST", strings.ToLower(b.id))
	default:
		return ""
	}
}

func (b WebEndpointBase) RestMappingBuilder(op Operation, group string,
	pathFunc MappingPathFunc, nameFunc MappingNameFunc) (*rest.MappingBuilder, error) {

	// NOTE: our current web implementation don't support different context-path (group)
	if group != "" {
		return nil, fmt.Errorf("adding actuator endpoints to different context-path/group is not supported at the moment")
	}

	path := pathFunc(op, b.properties)
	name := nameFunc(op)
	builder := rest.New(name).
		Path(path).
		EndpointFunc(op.Func())

	switch op.Mode() {
	case OperationRead:
		return builder.Method(http.MethodGet), nil
	case OperationWrite:
		return builder.Method(http.MethodPost), nil
	default:
		return nil, fmt.Errorf("unsupported operation mode")
	}
}

