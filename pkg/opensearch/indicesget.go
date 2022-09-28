package opensearch

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/opensearch-project/opensearch-go/opensearchapi"
	"io/ioutil"
	"net/http"
)

// IndicesDetail response follows opensearch spec
// [format] https://opensearch.org/docs/latest/opensearch/rest-api/index-apis/get-index/#response-body-fields
type IndicesDetail struct {
	Aliases  map[string]interface{} `json:"aliases"`
	Mappings map[string]interface{} `json:"mappings"`
	Settings struct {
		Index struct {
			CreationDate     string `json:"creation_date"`
			NumberOfShards   string `json:"number_of_shards"`
			NumberOfReplicas string `json:"number_of_replicas"`
			Uuid             string `json:"uuid"`
			Version          struct {
				Created string `json:"created"`
			} `json:"version"`
			ProvidedName string `json:"provided_name"`
		} `json:"index"`
	}
}

func (c *RepoImpl[T]) IndicesGet(ctx context.Context, index string, o ...Option[opensearchapi.IndicesGetRequest]) (*IndicesDetail, error) {

	resp, err := c.client.IndicesGet(ctx, index, o...)
	if resp.IsError() {
		return nil, fmt.Errorf("error status code: %d", resp.StatusCode)
	}
	if err != nil {
		return nil, err
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, fmt.Errorf("error status code: %d", resp.StatusCode)
	}

	indicesDetail := make(map[string]*IndicesDetail)
	err = json.Unmarshal(respBody, &indicesDetail)
	if err != nil {
		return nil, err
	}
	if len(indicesDetail) > 1 {
		return nil, fmt.Errorf("error status code: %d, more than one index exists, with the same alias/name: %s ", http.StatusInternalServerError, index)
	}
	// This is needed because the first level of the nested object returned will be an unknown index name (Assuming we use an alias)
	key := ""
	for k, _ := range indicesDetail {
		key += k
	}
	return indicesDetail[key], nil
}

func (c *OpenClientImpl) IndicesGet(ctx context.Context, index string, o ...Option[opensearchapi.IndicesGetRequest]) (*opensearchapi.Response, error) {
	options := make([]func(request *opensearchapi.IndicesGetRequest), len(o))
	for i, v := range o {
		options[i] = v
	}

	//nolint:makezero
	options = append(options, IndicesGet.WithContext(ctx))
	resp, err := c.client.Indices.Get([]string{index}, options...)

	return resp, err
}

type indicesGetExt struct {
	opensearchapi.IndicesGet
}

var IndicesGet = indicesGetExt{}
