package opensearch

import (
	"context"
	"fmt"
	"github.com/opensearch-project/opensearch-go/opensearchapi"
)

func (c *RepoImpl[T]) Ping(
	ctx context.Context,
	o ...Option[opensearchapi.PingRequest],
) error {
	resp, err := c.client.Ping(ctx, o...)
	if err != nil {
		return err
	}
	if resp != nil && resp.IsError() {
		return fmt.Errorf("error status code: %d", resp.StatusCode)
	}
	return nil
}

func (c *OpenClientImpl) Ping(ctx context.Context, o ...Option[opensearchapi.PingRequest]) (*opensearchapi.Response, error) {
	options := make([]func(request *opensearchapi.PingRequest), len(o))
	for i, v := range o {
		options[i] = v
	}
	for _, hook := range c.beforeHook {
		ctx = hook.Before(ctx, BeforeContext{cmd: CmdPing, Options: &options})
	}

	//nolint:makezero
	options = append(options, Ping.WithContext(ctx))
	resp, err := c.client.Ping(options...)

	for _, hook := range c.afterHook {
		ctx = hook.After(ctx, AfterContext{cmd: CmdPing, Options: &options, Resp: resp, Err: &err})
	}

	return resp, err
}

type pingExt struct {
	opensearchapi.Ping
}

var Ping = pingExt{}
