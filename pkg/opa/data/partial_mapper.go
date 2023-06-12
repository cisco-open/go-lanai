package opadata

import (
	"context"
	"encoding/json"
	"github.com/open-policy-agent/opa/rego"
)

type PartialQueryMapper struct {
	ctx context.Context
}

func (m *PartialQueryMapper) MapResults(pq *rego.PartialQueries) (interface{}, error) {
	_, _ = ParsePartialQueries(m.ctx, pq)
	return pq, nil
}

func (m *PartialQueryMapper) ResultToJSON(result interface{}) (interface{}, error) {
	data, e := json.Marshal(result)
	return string(data), e
}
