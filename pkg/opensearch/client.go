package opensearch

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/order"
	"errors"
	"fmt"
	"github.com/opensearch-project/opensearch-go"
	"github.com/opensearch-project/opensearch-go/opensearchapi"
	"go.uber.org/fx"
	"io"
	"reflect"
)

var (
	ErrCreatingNewClient = errors.New("unable to create opensearch client")
)

type Request interface {
	opensearchapi.SearchRequest |
		opensearchapi.IndicesCreateRequest |
		opensearchapi.IndexRequest |
		opensearchapi.IndicesDeleteRequest |
		opensearchapi.IndicesGetRequest |
		opensearchapi.IndicesPutAliasRequest |
		opensearchapi.IndicesDeleteAliasRequest |
		opensearchapi.IndicesPutIndexTemplateRequest |
		opensearchapi.IndicesDeleteIndexTemplateRequest |
		opensearchapi.PingRequest
}

type OpenClient interface {
	Search(ctx context.Context, o ...Option[opensearchapi.SearchRequest]) (*opensearchapi.Response, error)
	Index(ctx context.Context, index string, body io.Reader, o ...Option[opensearchapi.IndexRequest]) (*opensearchapi.Response, error)
	IndicesCreate(ctx context.Context, index string, o ...Option[opensearchapi.IndicesCreateRequest]) (*opensearchapi.Response, error)
	IndicesGet(ctx context.Context, index string, o ...Option[opensearchapi.IndicesGetRequest]) (*opensearchapi.Response, error)
	IndicesDelete(ctx context.Context, index string, o ...Option[opensearchapi.IndicesDeleteRequest]) (*opensearchapi.Response, error)
	IndicesPutAlias(ctx context.Context, index string, name string, o ...Option[opensearchapi.IndicesPutAliasRequest]) (*opensearchapi.Response, error)
	IndicesDeleteAlias(ctx context.Context, index string, name string, o ...Option[opensearchapi.IndicesDeleteAliasRequest]) (*opensearchapi.Response, error)
	IndicesPutIndexTemplate(ctx context.Context, name string, body io.Reader, o ...Option[opensearchapi.IndicesPutIndexTemplateRequest]) (*opensearchapi.Response, error)
	IndicesDeleteIndexTemplate(ctx context.Context, name string, o ...Option[opensearchapi.IndicesDeleteIndexTemplateRequest]) (*opensearchapi.Response, error)
	Ping(ctx context.Context, o ...Option[opensearchapi.PingRequest]) (*opensearchapi.Response, error)
	AddBeforeHook(hook BeforeHook)
	AddAfterHook(hook AfterHook)
	RemoveBeforeHook(hook BeforeHook)
	RemoveAfterHook(hook AfterHook)
}

type Option[T Request] func(request *T)

const (
	// FxGroup defines the FX group for the OpenSearch
	FxGroup = "opensearch"
)

type newClientDI struct {
	fx.In
	Config opensearch.Config

	BeforeHook []BeforeHook `group:"opensearch"`
	AfterHook  []AfterHook  `group:"opensearch"`
}

func NewClient(di newClientDI) (OpenClient, error) {
	client, err := opensearch.NewClient(di.Config)
	if err != nil {
		return nil, fmt.Errorf("%w, %v", ErrCreatingNewClient, err)
	}
	order.SortStable(di.BeforeHook, order.OrderedFirstCompare)
	order.SortStable(di.AfterHook, order.OrderedFirstCompare)
	openClientImpl := &OpenClientImpl{
		client:     client,
		beforeHook: di.BeforeHook,
		afterHook:  di.AfterHook,
	}

	return openClientImpl, nil
}

type configDI struct {
	fx.In
	Properties *Properties
}

func NewConfig(di configDI) opensearch.Config {
	return opensearch.Config{
		Addresses: di.Properties.Addresses,
		Username:  di.Properties.Username,
		Password:  di.Properties.Password,
	}
}

type OpenClientImpl struct {
	client     *opensearch.Client
	beforeHook []BeforeHook
	afterHook  []AfterHook
}

// CommandType lets the hooks know what command is being run
type CommandType int

const (
	UnknownCommand string = "unknown"
)
const (
	CmdSearch CommandType = iota
	CmdIndex
	CmdIndicesCreate
	CmdIndicesGet
	CmdIndicesDelete
	CmdIndicesPutAlias
	CmdIndicesDeleteAlias
	CmdIndicesPutIndexTemplate
	CmdIndicesDeleteIndexTemplate
	CmdPing
)

var CmdToString = map[CommandType]string{
	CmdSearch:                     "search",
	CmdIndex:                      "index",
	CmdIndicesCreate:              "indices create",
	CmdIndicesGet:                 "indices get",
	CmdIndicesDelete:              "indices delete",
	CmdIndicesPutAlias:            "indices put alias",
	CmdIndicesDeleteAlias:         "indices delete alias",
	CmdIndicesPutIndexTemplate:    "indices put index template",
	CmdIndicesDeleteIndexTemplate: "indices delete index template",
	CmdPing:                       "ping",
}

// String will return the command in string format. If the command is not found
// the UnknownCommand string will be returned
func (c CommandType) String() string {
	val, ok := CmdToString[c]
	if !ok {
		logger.Errorf("unknown command: %v", c)
		return UnknownCommand
	}
	return val
}

func (c *OpenClientImpl) AddBeforeHook(hook BeforeHook) {
	c.beforeHook = append(c.beforeHook, hook)
	order.SortStable(c.beforeHook, order.OrderedFirstCompare)
}

func (c *OpenClientImpl) AddAfterHook(hook AfterHook) {
	c.afterHook = append(c.afterHook, hook)
	order.SortStable(c.afterHook, order.OrderedFirstCompare)
}

// RemoveBeforeHook will remove the given BeforeHook. To ensure your hook is removable,
// the hook should implement the Identifier interface. If not, your hooks should be
// distinct in the eyes of reflect.DeepEqual, otherwise the hook will not be removed.
func (c *OpenClientImpl) RemoveBeforeHook(hook BeforeHook) {
	if hookWithIdentifier, ok := hook.(Identifier); ok {
		for i, beforeHook := range c.beforeHook {
			if beforeHookWithIdentifier, hok := beforeHook.(Identifier); hok {
				if hookWithIdentifier.ID() == beforeHookWithIdentifier.ID() {
					c.beforeHook = utils.RemoveStable(c.beforeHook, i)
				}
			}
		}
		return
	}

	for i, h := range c.beforeHook {
		if reflect.DeepEqual(h, hook) {
			c.beforeHook = utils.RemoveStable(c.beforeHook, i)
		}
	}
}

// RemoveAfterHook will remove the given AfterHook. To ensure your hook is removable,
// the hook should implement the Identifier interface. If not, your hooks should be
// distinct in the eyes of reflect.DeepEqual, otherwise the hook will not be removed.
func (c *OpenClientImpl) RemoveAfterHook(hook AfterHook) {
	if hookWithIdentifier, ok := hook.(Identifier); ok {
		for i, afterHook := range c.afterHook {
			if afterHookWithIdentifier, hok := afterHook.(Identifier); hok {
				if hookWithIdentifier.ID() == afterHookWithIdentifier.ID() {
					c.afterHook = utils.RemoveStable(c.afterHook, i)
				}
			}
		}
		return
	}

	for i, h := range c.afterHook {
		if reflect.DeepEqual(h, hook) {
			c.afterHook = utils.RemoveStable(c.afterHook, i)
		}
	}
}

// BeforeContext is the context given to a BeforeHook
//
// Options will be in the form *[]func(request *Request){}, example:
// 	options := make([]func(request *opensearchapi.SearchRequest), 0)
// 	BeforeContext{Options: &options}
type BeforeContext struct {
	cmd     CommandType
	Options interface{}
}

func (c *BeforeContext) CommandType() CommandType {
	return c.cmd
}

type BeforeHook interface {
	Before(ctx context.Context, before BeforeContext) context.Context
}

type Identifier interface {
	ID() string
}

// BeforeHookBase provides a way to create an BeforeHook, similar to BeforeHookFunc,
// but in a way that implements the Identifier interface so that it can be removed using
// the RemoveBeforeHook function
type BeforeHookBase struct {
	Identifier string
	F          func(ctx context.Context, after BeforeContext) context.Context
}

func (s BeforeHookBase) Before(ctx context.Context, before BeforeContext) context.Context {
	return s.F(ctx, before)
}

func (s BeforeHookBase) ID() string {
	return s.Identifier
}

type BeforeHookFunc func(ctx context.Context, before BeforeContext) context.Context

func (f BeforeHookFunc) Before(ctx context.Context, before BeforeContext) context.Context {
	return f(ctx, before)
}

// AfterContext is the context given to a AfterHook
//
// Options will be in the form *[]func(request *Request){} example:
// 	options := make([]func(request *opensearchapi.SearchRequest), 0)
// 	AfterContext{Options: &options}
//
// Resp and Err can be modified before they are returned out of Request of OpenClientImpl
// example being OpenClientImpl.Search
type AfterContext struct {
	cmd     CommandType
	Options interface{}
	Resp    *opensearchapi.Response
	Err     *error
}

func (c *AfterContext) CommandType() CommandType {
	return c.cmd
}

type AfterHook interface {
	After(ctx context.Context, after AfterContext) context.Context
}

// AfterHookBase provides a way to create an AfterHook, similar to AfterHookFunc.
// but in a way that implements the Identifier interface so that it can be removed using
// the RemoveAfterHook function
type AfterHookBase struct {
	Identifier string
	F          func(ctx context.Context, after AfterContext) context.Context
}

func (s AfterHookBase) After(ctx context.Context, after AfterContext) context.Context {
	return s.F(ctx, after)
}

func (s AfterHookBase) ID() string {
	return s.Identifier
}

// AfterHookFunc provides a way to easily create a AfterHook - however hooks created in this
// manner are not able to be deleted from the hook slice
type AfterHookFunc func(ctx context.Context, after AfterContext) context.Context

func (f AfterHookFunc) After(ctx context.Context, after AfterContext) context.Context {
	return f(ctx, after)
}
