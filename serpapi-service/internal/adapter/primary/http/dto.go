package http

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/JIeeiroSst/serpapi-service/internal/domain"
)

type searchRequestDTO struct {
	Engine  string                     `json:"engine"`
	Params  map[string]json.RawMessage `json:"params"`
	Output  string                     `json:"output,omitempty"`
	NoCache bool                       `json:"no_cache,omitempty"`
	Async   bool                       `json:"async,omitempty"`
}

func (d searchRequestDTO) toDomain() (domain.SearchRequest, error) {
	out, err := domain.ParseOutput(d.Output)
	if err != nil {
		return domain.SearchRequest{}, err
	}
	params := make(domain.Params, len(d.Params))
	for k, raw := range d.Params {
		v, err := scalar(raw)
		if err != nil {
			return domain.SearchRequest{}, domain.Invalid("params.%s: %v", k, err)
		}
		params[k] = v
	}
	return domain.SearchRequest{
		Engine:  d.Engine,
		Params:  params,
		Output:  out,
		NoCache: d.NoCache,
		Async:   d.Async,
	}, nil
}

func scalar(raw json.RawMessage) (string, error) {
	var v any
	if err := json.Unmarshal(raw, &v); err != nil {
		return "", err
	}
	switch t := v.(type) {
	case string:
		return t, nil
	case bool:
		return strconv.FormatBool(t), nil
	case float64:
		return strconv.FormatFloat(t, 'f', -1, 64), nil
	}
	return "", fmt.Errorf("must be a string, number or boolean")
}

type batchRequestDTO struct {
	Searches []searchRequestDTO `json:"searches"`
}

type batchItemDTO struct {
	Status   int             `json:"status"`
	SearchID string          `json:"search_id,omitempty"`
	Cached   bool            `json:"cached,omitempty"`
	Error    string          `json:"error,omitempty"`
	Data     json.RawMessage `json:"data,omitempty"`
}

type batchResponseDTO struct {
	Results []batchItemDTO `json:"results"`
}
