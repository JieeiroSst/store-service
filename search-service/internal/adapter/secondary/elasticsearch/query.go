package elasticsearch

import (
	"fmt"
	"sort"
	"strings"

	"github.com/JIeeiroSst/search-service/internal/domain"
)

const (
	allTextField   = "all_text"
	facetSize      = 20
	indexFacetSize = 100
)

var numericTypes = map[string]bool{
	"long": true, "integer": true, "short": true, "byte": true,
	"double": true, "float": true, "half_float": true, "scaled_float": true, "unsigned_long": true,
}

type fieldCap struct {
	types        []string
	searchable   bool
	aggregatable bool
}

type resolvedField struct {
	name        string
	typ         string
	numericType string
}

func capsCandidates(q domain.DocumentQuery) []string {
	var fields []string
	add := func(f string) {
		fields = append(fields, f, f+".keyword", f+".sort")
	}
	for _, s := range q.Sort {
		if s.Field != domain.ScoreField {
			add(s.Field)
		}
	}
	for _, f := range q.Facets {
		if f != domain.IndexFacet {
			add(f)
		}
	}
	return fields
}

func resolveAggregatable(caps map[string]fieldCap, field string, candidates []string) (resolvedField, error) {
	found := false
	for _, name := range candidates {
		c, ok := caps[name]
		if !ok {
			continue
		}
		found = true
		if len(c.types) == 1 {
			if c.aggregatable {
				return resolvedField{name: name, typ: c.types[0]}, nil
			}
			continue
		}
		switch {
		case !c.aggregatable:
		case allOf(c.types, func(t string) bool { return numericTypes[t] }):
			return resolvedField{name: name, typ: "double", numericType: "double"}, nil
		case allOf(c.types, func(t string) bool { return t == "date" || t == "date_nanos" }):
			return resolvedField{name: name, typ: "date", numericType: "date_nanos"}, nil
		}
		return resolvedField{}, domain.Invalid("field %q has conflicting types %v across indices, narrow the index parameter", name, c.types)
	}
	if !found {
		return resolvedField{}, domain.Invalid("field %q does not exist in the selected indices", field)
	}
	return resolvedField{}, domain.Invalid("field %q cannot be used for sorting or facets", field)
}

func allOf(values []string, pred func(string) bool) bool {
	for _, v := range values {
		if !pred(v) {
			return false
		}
	}
	return true
}

func buildQuery(q domain.DocumentQuery) map[string]any {
	boolQuery := map[string]any{}
	if q.Keyword != "" {
		boolQuery["must"] = keywordQuery(q.Keyword)
	}

	var filters []any
	for _, field := range sortedKeys(q.Filters) {
		should := make([]any, 0, len(q.Filters[field]))
		for _, v := range q.Filters[field] {
			should = append(should, map[string]any{
				"match": map[string]any{
					field: map[string]any{"query": v, "operator": "and", "lenient": true},
				},
			})
		}
		filters = append(filters, map[string]any{
			"bool": map[string]any{"should": should, "minimum_should_match": 1},
		})
	}
	for _, field := range sortedKeys(q.Ranges) {
		filters = append(filters, map[string]any{
			"range": map[string]any{field: q.Ranges[field]},
		})
	}
	if len(filters) > 0 {
		boolQuery["filter"] = filters
	}

	if len(boolQuery) == 0 {
		return map[string]any{"match_all": map[string]any{}}
	}
	return map[string]any{"bool": boolQuery}
}

func keywordQuery(keyword string) map[string]any {
	return map[string]any{
		"bool": map[string]any{
			"should": []any{
				map[string]any{"match_phrase": map[string]any{
					allTextField: map[string]any{"query": keyword, "slop": 2, "boost": 4},
				}},
				map[string]any{"match": map[string]any{
					allTextField: map[string]any{"query": keyword, "operator": "and", "boost": 3},
				}},
				map[string]any{"match": map[string]any{
					allTextField + ".autocomplete": map[string]any{"query": keyword, "operator": "and", "boost": 1.5},
				}},
				map[string]any{"match": map[string]any{
					allTextField: map[string]any{"query": keyword, "operator": "and", "fuzziness": "AUTO", "prefix_length": 1},
				}},
			},
			"minimum_should_match": 1,
		},
	}
}

func autocompleteQuery(keyword string) map[string]any {
	return map[string]any{
		"bool": map[string]any{
			"must": map[string]any{"match": map[string]any{
				allTextField + ".autocomplete": map[string]any{"query": keyword, "operator": "and"},
			}},
			"should": map[string]any{"match_phrase_prefix": map[string]any{
				allTextField: map[string]any{"query": keyword, "boost": 2},
			}},
		},
	}
}

func buildSort(q domain.DocumentQuery, caps map[string]fieldCap) ([]any, error) {
	var out []any
	for _, s := range q.Sort {
		order := "asc"
		if s.Desc {
			order = "desc"
		}
		if s.Field == domain.ScoreField {
			out = append(out, map[string]any{domain.ScoreField: order})
			continue
		}
		field, err := resolveAggregatable(caps, s.Field, []string{s.Field, s.Field + ".sort", s.Field + ".keyword"})
		if err != nil {
			return nil, err
		}
		clause := map[string]any{"order": order, "missing": "_last", "unmapped_type": field.typ}
		if field.numericType != "" {
			clause["numeric_type"] = field.numericType
		}
		out = append(out, map[string]any{field.name: clause})
	}
	if len(q.Sort) == 0 && q.Keyword != "" {
		out = append(out, map[string]any{domain.ScoreField: "desc"})
	}
	return append(out,
		map[string]any{"_index": "asc"},
		map[string]any{"_id": "asc"},
	), nil
}

func buildFacets(facets []string, caps map[string]fieldCap) (map[string]any, error) {
	aggs := map[string]any{
		domain.IndexFacet: map[string]any{
			"terms": map[string]any{"field": "_index", "size": indexFacetSize},
		},
	}
	for _, f := range facets {
		if f == domain.IndexFacet {
			continue
		}
		field, err := resolveAggregatable(caps, f, []string{f, f + ".keyword", f + ".sort"})
		if err != nil {
			return nil, err
		}
		aggs[f] = map[string]any{
			"terms": map[string]any{"field": field.name, "size": facetSize},
		}
	}
	return aggs, nil
}

func highlightClause() map[string]any {
	return map[string]any{
		"fields":              map[string]any{"*": map[string]any{}},
		"require_field_match": false,
		"number_of_fragments": 3,
		"fragment_size":       150,
		"pre_tags":            []string{"<em>"},
		"post_tags":           []string{"</em>"},
	}
}

func autocompleteHighlight(keyword string) map[string]any {
	return map[string]any{
		"fields":              map[string]any{"*": map[string]any{}},
		"require_field_match": false,
		"number_of_fragments": 1,
		"fragment_size":       80,
		"highlight_query": map[string]any{
			"multi_match": map[string]any{
				"query":   keyword,
				"type":    "phrase_prefix",
				"fields":  []string{"*"},
				"lenient": true,
			},
		},
	}
}

func suggestClause(keyword string) map[string]any {
	return map[string]any{
		"text": keyword,
		"did_you_mean": map[string]any{
			"phrase": map[string]any{
				"field":      allTextField + ".shingle",
				"size":       1,
				"gram_size":  3,
				"confidence": 1,
				"direct_generator": []any{
					map[string]any{"field": allTextField + ".shingle", "suggest_mode": "missing"},
				},
			},
		},
	}
}

func similarQuery(ref domain.DocumentRef) map[string]any {
	return map[string]any{
		"more_like_this": map[string]any{
			"like":                 []any{map[string]any{"_index": ref.Index, "_id": ref.ID}},
			"min_term_freq":        1,
			"min_doc_freq":         1,
			"max_query_terms":      25,
			"minimum_should_match": 1,
		},
	}
}

func toFacets(res searchResponse) map[string][]domain.FacetBucket {
	if len(res.Aggregations) == 0 {
		return nil
	}
	facets := make(map[string][]domain.FacetBucket, len(res.Aggregations))
	for name, agg := range res.Aggregations {
		buckets := make([]domain.FacetBucket, 0, len(agg.Buckets))
		for _, b := range agg.Buckets {
			var value any = b.Key
			if b.KeyAsString != "" {
				value = b.KeyAsString
			}
			if name == domain.IndexFacet {
				value = displayIndex(fmt.Sprint(b.Key))
			}
			buckets = append(buckets, domain.FacetBucket{Value: value, Count: b.DocCount})
		}
		facets[name] = buckets
	}
	return facets
}

func toSuggestion(res searchResponse, keyword string) string {
	entries := res.Suggest["did_you_mean"]
	if len(entries) == 0 || len(entries[0].Options) == 0 {
		return ""
	}
	text := entries[0].Options[0].Text
	if strings.EqualFold(text, keyword) {
		return ""
	}
	return text
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
