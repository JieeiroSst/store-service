package elasticsearch

import (
	"context"
	"sort"
	"strconv"
	"strings"

	"github.com/JIeeiroSst/search-service/config"
	"github.com/JIeeiroSst/search-service/internal/domain"
	esv7 "github.com/elastic/go-elasticsearch/v7"
	esv7api "github.com/elastic/go-elasticsearch/v7/esapi"
)

type searchHit struct {
	Index     string              `json:"_index"`
	ID        string              `json:"_id"`
	Score     *float64            `json:"_score"`
	Source    map[string]any      `json:"_source"`
	Highlight map[string][]string `json:"highlight"`
	Sort      []any               `json:"sort"`
}

type searchResponse struct {
	Hits struct {
		Total struct {
			Value int64 `json:"value"`
		} `json:"total"`
		Hits []searchHit `json:"hits"`
	} `json:"hits"`
	Aggregations map[string]struct {
		Buckets []struct {
			Key         any    `json:"key"`
			KeyAsString string `json:"key_as_string"`
			DocCount    int64  `json:"doc_count"`
		} `json:"buckets"`
	} `json:"aggregations"`
	Suggest map[string][]struct {
		Options []struct {
			Text string `json:"text"`
		} `json:"options"`
	} `json:"suggest"`
}

type Searcher struct {
	client  *esv7.Client
	indices []string
}

func NewSearcher(client *esv7.Client, cfg *config.Config) *Searcher {
	indices := cfg.Elasticsearch.Indices
	if len(indices) == 0 {
		indices = defaultIndices
	}
	return &Searcher{client: client, indices: indices}
}

func (s *Searcher) Query(ctx context.Context, q domain.DocumentQuery, mode domain.QueryMode) (domain.DocumentPage, error) {
	indices := s.resolveIndices(q.Indices)
	kind := string(mode)
	fingerprint := fingerprintOf(kind, q.Fingerprint(), indices)
	cur, err := decodeCursor(q.Cursor, kind, fingerprint)
	if err != nil {
		return domain.DocumentPage{}, err
	}

	caps, err := s.fieldCaps(ctx, indices, capsCandidates(q))
	if err != nil {
		return domain.DocumentPage{}, err
	}
	sortClause, err := buildSort(q, caps)
	if err != nil {
		return domain.DocumentPage{}, err
	}

	body := map[string]any{
		"query":        buildQuery(q),
		"sort":         sortClause,
		"track_scores": q.Keyword != "",
		"size":         q.Size + 1,
	}
	if q.Keyword != "" {
		body["highlight"] = highlightClause()
	}
	if cur != nil {
		body["search_after"] = cur.SearchAfter
	} else {
		aggs, err := buildFacets(q.Facets, caps)
		if err != nil {
			return domain.DocumentPage{}, err
		}
		body["aggs"] = aggs
		if q.Keyword != "" {
			body["suggest"] = suggestClause(q.Keyword)
		}
	}

	var res searchResponse
	if err := s.search(ctx, indices, body, &res); err != nil {
		return domain.DocumentPage{}, err
	}

	hits := res.Hits.Hits
	hasMore := len(hits) > q.Size
	if hasMore {
		hits = hits[:q.Size]
	}

	page := domain.DocumentPage{
		Items:      toDocuments(hits),
		Total:      res.Hits.Total.Value,
		HasMore:    hasMore,
		Facets:     toFacets(res),
		Suggestion: toSuggestion(res, q.Keyword),
	}
	if hasMore {
		next, err := encodeCursor(cursor{Kind: kind, Fingerprint: fingerprint, SearchAfter: hits[len(hits)-1].Sort})
		if err != nil {
			return domain.DocumentPage{}, err
		}
		page.NextCursor = next
	}
	return page, nil
}

func (s *Searcher) Autocomplete(ctx context.Context, q domain.AutocompleteQuery) ([]domain.Document, error) {
	body := map[string]any{
		"query":     autocompleteQuery(q.Keyword),
		"size":      q.Size,
		"highlight": autocompleteHighlight(q.Keyword),
	}
	var res searchResponse
	if err := s.search(ctx, s.resolveIndices(q.Indices), body, &res); err != nil {
		return nil, err
	}
	return toDocuments(res.Hits.Hits), nil
}

func (s *Searcher) Get(ctx context.Context, ref domain.DocumentRef) (domain.Document, error) {
	var hit searchHit
	if err := call(ctx, s.client, esv7api.GetRequest{Index: ref.Index, DocumentID: ref.ID}, &hit); err != nil {
		return domain.Document{}, err
	}
	return toDocuments([]searchHit{hit})[0], nil
}

func (s *Searcher) Similar(ctx context.Context, ref domain.DocumentRef, size int) ([]domain.Document, error) {
	if _, err := s.Get(ctx, ref); err != nil {
		return nil, err
	}
	var res searchResponse
	if err := s.search(ctx, s.indices, map[string]any{"query": similarQuery(ref), "size": size}, &res); err != nil {
		return nil, err
	}
	return toDocuments(res.Hits.Hits), nil
}

func (s *Searcher) Indices(ctx context.Context) ([]domain.IndexInfo, error) {
	var rows []struct {
		Index     string `json:"index"`
		DocsCount string `json:"docs.count"`
	}
	err := call(ctx, s.client, esv7api.CatIndicesRequest{
		Index:           s.indices,
		Format:          "json",
		H:               []string{"index", "docs.count"},
		S:               []string{"index"},
		ExpandWildcards: "open",
	}, &rows)
	if err != nil {
		return nil, err
	}

	var mappings map[string]struct {
		Mappings map[string]any `json:"mappings"`
	}
	err = call(ctx, s.client, esv7api.IndicesGetFieldMappingRequest{
		Index:             s.indices,
		Fields:            []string{allTextField},
		ExpandWildcards:   "open",
		IgnoreUnavailable: boolPtr(true),
		AllowNoIndices:    boolPtr(true),
	}, &mappings)
	if err != nil {
		return nil, err
	}

	var settings map[string]struct {
		Settings map[string]string `json:"settings"`
	}
	err = call(ctx, s.client, esv7api.IndicesGetSettingsRequest{
		Index:             s.indices,
		Name:              []string{"index.blocks.write"},
		FlatSettings:      boolPtr(true),
		ExpandWildcards:   "open",
		IgnoreUnavailable: boolPtr(true),
		AllowNoIndices:    boolPtr(true),
	}, &settings)
	if err != nil {
		return nil, err
	}

	res := make([]domain.IndexInfo, 0, len(rows))
	for _, row := range rows {
		count, _ := strconv.ParseInt(row.DocsCount, 10, 64)
		_, managed := mappings[row.Index].Mappings[allTextField]
		res = append(res, domain.IndexInfo{
			Name:         displayIndex(row.Index),
			Index:        row.Index,
			DocsCount:    count,
			Managed:      managed,
			WriteBlocked: settings[row.Index].Settings["index.blocks.write"] == "true",
		})
	}
	return res, nil
}

func (s *Searcher) Fields(ctx context.Context, index string) ([]domain.FieldInfo, error) {
	caps, err := s.fieldCaps(ctx, []string{index}, []string{"*"})
	if err != nil {
		return nil, err
	}
	res := make([]domain.FieldInfo, 0, len(caps))
	for name, c := range caps {
		if strings.HasPrefix(name, "_") || strings.HasPrefix(name, allTextField) || c.types[0] == "object" || c.types[0] == "nested" {
			continue
		}
		res = append(res, domain.FieldInfo{Name: name, Types: c.types, Searchable: c.searchable, Aggregatable: c.aggregatable})
	}
	sort.Slice(res, func(i, j int) bool { return res[i].Name < res[j].Name })
	return res, nil
}

func (s *Searcher) resolveIndices(requested []string) []string {
	if len(requested) == 0 {
		return s.indices
	}
	return append(append([]string{}, requested...), "-.*")
}

func (s *Searcher) search(ctx context.Context, indices []string, body map[string]any, out *searchResponse) error {
	reader, err := jsonBody(body)
	if err != nil {
		return err
	}
	return call(ctx, s.client, esv7api.SearchRequest{
		Index:             indices,
		Body:              reader,
		IgnoreUnavailable: boolPtr(true),
		AllowNoIndices:    boolPtr(true),
		ExpandWildcards:   "open",
		TrackTotalHits:    true,
	}, out)
}

func (s *Searcher) fieldCaps(ctx context.Context, indices, fields []string) (map[string]fieldCap, error) {
	if len(fields) == 0 {
		return map[string]fieldCap{}, nil
	}
	var res struct {
		Fields map[string]map[string]struct {
			Searchable   bool `json:"searchable"`
			Aggregatable bool `json:"aggregatable"`
		} `json:"fields"`
	}
	err := call(ctx, s.client, esv7api.FieldCapsRequest{
		Index:             indices,
		Fields:            fields,
		ExpandWildcards:   "open",
		IgnoreUnavailable: boolPtr(true),
		AllowNoIndices:    boolPtr(true),
	}, &res)
	if err != nil {
		return nil, err
	}

	caps := make(map[string]fieldCap, len(res.Fields))
	for name, byType := range res.Fields {
		c := fieldCap{searchable: true, aggregatable: true}
		for typ, info := range byType {
			if typ == "unmapped" {
				continue
			}
			c.types = append(c.types, typ)
			c.searchable = c.searchable && info.Searchable
			c.aggregatable = c.aggregatable && info.Aggregatable
		}
		if len(c.types) == 0 {
			continue
		}
		sort.Strings(c.types)
		caps[name] = c
	}
	return caps, nil
}

func toDocuments(hits []searchHit) []domain.Document {
	items := make([]domain.Document, len(hits))
	for i, hit := range hits {
		for field := range hit.Highlight {
			if strings.HasPrefix(field, allTextField) {
				delete(hit.Highlight, field)
			}
		}
		if len(hit.Highlight) == 0 {
			hit.Highlight = nil
		}
		items[i] = domain.Document{
			Index:     displayIndex(hit.Index),
			ID:        hit.ID,
			Score:     hit.Score,
			Source:    hit.Source,
			Highlight: hit.Highlight,
		}
	}
	return items
}
