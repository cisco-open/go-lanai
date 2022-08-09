package opensearch

import (
	"context"
	"errors"
	"fmt"
	"github.com/opensearch-project/opensearch-go"
	"github.com/opensearch-project/opensearch-go/opensearchapi"
	"go.uber.org/fx"
	"io"
)

var (
	ErrCreatingNewClient = errors.New("unable to create opensearch client")
)

type Request interface {
	opensearchapi.SearchRequest |
		opensearchapi.IndicesCreateRequest |
		opensearchapi.IndexRequest |
		opensearchapi.IndicesDeleteRequest
}

type OpenClient interface {
	Search(ctx context.Context, o ...Option[opensearchapi.SearchRequest]) (*opensearchapi.Response, error)
	Index(ctx context.Context, index string, body io.Reader, o ...Option[opensearchapi.IndexRequest]) (*opensearchapi.Response, error)
	IndicesCreate(ctx context.Context, index string, o ...Option[opensearchapi.IndicesCreateRequest]) (*opensearchapi.Response, error)
	IndicesDelete(ctx context.Context, index string, o ...Option[opensearchapi.IndicesDeleteRequest]) (*opensearchapi.Response, error)
	AddHook(hook HookContainer)
}

type Option[T Request] func(request *T)

const (
	// FxOpenSearchHooksGroup defines the FX group for the OpenSearch hooks
	FxOpenSearchHooksGroup = "opensearchhooks"
)

type newClientDI struct {
	fx.In
	Hooks  []HookContainer `group:"opensearchhooks"`
	Config opensearch.Config
}

func NewClient(di newClientDI) (OpenClient, error) {
	client, err := opensearch.NewClient(di.Config)
	if err != nil {
		return nil, fmt.Errorf("%w, %v", ErrCreatingNewClient, err)
	}
	return &OpenClientImpl{
		client:         client,
		hookContainers: di.Hooks,
	}, nil
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
	client         *opensearch.Client
	hookContainers []HookContainer
}

// CommandType lets the hooks know what command is being run
type CommandType int

const (
	CmdSearch CommandType = iota
	CmdIndex
	CmdIndicesCreate
	CmdIndicesDelete
)

// HookContext contains any context we may need for the hooks
type HookContext struct {
	Ctx context.Context
	Cmd CommandType
}
type HookContainer struct {
	Before func(ctx HookContext)
	After  func(ctx HookContext)
}

func (c *OpenClientImpl) AddHook(hook HookContainer) {
	c.hookContainers = append(c.hookContainers, hook)
}

// GetHooks will return only the non-nil hooks
func (c *OpenClientImpl) GetHooks() (before Hooks, after Hooks) {
	for _, hook := range c.hookContainers {
		if hook.Before != nil {
			before = append(before, hook.Before)
		}
		if hook.After != nil {
			after = append(after, hook.After)
		}
	}
	return
}

type Hooks []func(ctx HookContext)

func (h Hooks) Run(hookContext HookContext) {
	for _, hook := range h {
		hook(hookContext)
	}
}
