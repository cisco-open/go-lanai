package opensearch

import (
	"bytes"
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/order"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/opensearch-project/opensearch-go/opensearchapi"
	"io"
	"net/http"
)

var (
	ErrIndexNotFound = errors.New("index not found")
)

// SearchResponse modeled after https://opensearch.org/docs/latest/opensearch/rest-api/search/#response-body
type SearchResponse[T any] struct {
	Took     int  `json:"took"`
	TimedOut bool `json:"timed_out"`
	Shards   struct {
		Total      int `json:"total"`
		Successful int `json:"successful"`
		Skipped    int `json:"skipped"`
		Failed     int `json:"failed"`
	}
	Hits struct {
		MaxScore float64 `json:"max_score"`
		Total    struct {
			Value int `json:"value"`
		} `json:"total"`
		Hits []struct {
			Index  string  `json:"_index"`
			ID     string  `json:"_id"`
			Score  float64 `json:"_score"`
			Source T       `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
}

func (c *RepoImpl[T]) Search(ctx context.Context, dest *[]T, body interface{}, o ...Option[opensearchapi.SearchRequest]) (hits int, err error) {
	var buffer bytes.Buffer
	err = json.NewEncoder(&buffer).Encode(body)
	if err != nil {
		return 0, fmt.Errorf("unable to encode mapping: %w", err)
	}
	o = append(o, Search.WithBody(&buffer))
	resp, err := c.client.Search(ctx, o...)
	if err != nil {
		return 0, err
	}
	if resp != nil && resp.IsError() {
		logger.WithContext(ctx).Errorf("error response: %s", resp.String())
		if resp.StatusCode == http.StatusNotFound {
			return 0, fmt.Errorf("%w", ErrIndexNotFound)
		} else {
			return 0, fmt.Errorf("error status code: %d", resp.StatusCode)
		}
	}
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}
	var searchResp SearchResponse[T]
	err = json.Unmarshal(respBody, &searchResp)
	if err != nil {
		return 0, err
	}
	retModel := make([]T, len(searchResp.Hits.Hits))
	for i, hits := range searchResp.Hits.Hits {
		retModel[i] = hits.Source
	}
	*dest = retModel
	return searchResp.Hits.Total.Value, nil
}

func (c *OpenClientImpl) Search(ctx context.Context, o ...Option[opensearchapi.SearchRequest]) (*opensearchapi.Response, error) {
	options := make([]func(request *opensearchapi.SearchRequest), len(o))
	for i, v := range o {
		options[i] = v
	}

	order.SortStable(c.beforeHook, order.OrderedFirstCompare)
	for _, hook := range c.beforeHook {
		ctx = hook.Before(ctx, BeforeContext{cmd: CmdSearch, Options: &options})
	}

	//nolint:makezero
	options = append(options, Search.WithContext(ctx))
	resp, err := c.client.API.Search(options...)

	for _, hook := range c.afterHook {
		ctx = hook.After(ctx, AfterContext{cmd: CmdSearch, Options: &options, Resp: resp, Err: &err})
	}

	return resp, err
}

// searchExt can be extended
//
//	func (s searchExt) WithSomething() func(request *opensearchapi.SearchRequest) {
//		return func(request *opensearchapi.SearchRequest) {
//		}
//	}
type searchExt struct {
	opensearchapi.Search
}

var Search = searchExt{}
