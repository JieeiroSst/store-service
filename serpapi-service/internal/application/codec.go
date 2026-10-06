package application

import (
	"bytes"
	"encoding/json"
	"errors"

	"github.com/JIeeiroSst/serpapi-service/internal/domain"
)

type resultHeader struct {
	ContentType string `json:"c"`
	SearchID    string `json:"i,omitempty"`
	Status      string `json:"s,omitempty"`
}

func encodeResult(r domain.SearchResult) []byte {
	h, _ := json.Marshal(resultHeader{ContentType: r.ContentType, SearchID: r.SearchID, Status: r.Status})
	out := make([]byte, 0, len(h)+1+len(r.Body))
	out = append(out, h...)
	out = append(out, '\n')
	return append(out, r.Body...)
}

func decodeResult(b []byte) (domain.SearchResult, error) {
	i := bytes.IndexByte(b, '\n')
	if i < 0 {
		return domain.SearchResult{}, errors.New("malformed cache entry")
	}
	var h resultHeader
	if err := json.Unmarshal(b[:i], &h); err != nil {
		return domain.SearchResult{}, err
	}
	return domain.SearchResult{
		Body:        b[i+1:],
		ContentType: h.ContentType,
		SearchID:    h.SearchID,
		Status:      h.Status,
	}, nil
}
