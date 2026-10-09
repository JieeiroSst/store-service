package domain

import (
	"encoding/json"
	"regexp"
	"strings"
)

const (
	DefaultPageSize         = 20
	MaxPageSize             = 100
	DefaultAutocompleteSize = 10
	MaxAutocompleteSize     = 20
	DefaultSimilarSize      = 10
	MaxKeywordLen           = 256
	MaxSortFields           = 3
	MaxFacets               = 5
	MaxFilters              = 10
	IndexFacet              = "_index"
	ScoreField              = "_score"
)

var (
	fieldNamePattern = regexp.MustCompile(`^[A-Za-z0-9@][A-Za-z0-9_.@-]{0,127}$`)
	indexNamePattern = regexp.MustCompile(`^[a-z0-9][a-z0-9._*-]{0,254}$`)
)

type QueryMode string

const (
	ModeSearch QueryMode = "search"
	ModeList   QueryMode = "list"
)

type Document struct {
	Index     string              `json:"index"`
	ID        string              `json:"id"`
	Score     *float64            `json:"score,omitempty"`
	Source    map[string]any      `json:"source"`
	Highlight map[string][]string `json:"highlight,omitempty"`
}

type FacetBucket struct {
	Value any   `json:"value"`
	Count int64 `json:"count"`
}

type DocumentPage struct {
	Items      []Document               `json:"items"`
	Total      int64                    `json:"total"`
	NextCursor string                   `json:"next_cursor,omitempty"`
	HasMore    bool                     `json:"has_more"`
	Facets     map[string][]FacetBucket `json:"facets,omitempty"`
	Suggestion string                   `json:"suggestion,omitempty"`
}

type IndexInfo struct {
	Name         string `json:"name"`
	Index        string `json:"index"`
	DocsCount    int64  `json:"docs_count"`
	Managed      bool   `json:"managed"`
	WriteBlocked bool   `json:"write_blocked"`
}

type FieldInfo struct {
	Name         string   `json:"name"`
	Types        []string `json:"types"`
	Searchable   bool     `json:"searchable"`
	Aggregatable bool     `json:"aggregatable"`
}

type Range struct {
	Gt  string `json:"gt,omitempty"`
	Gte string `json:"gte,omitempty"`
	Lt  string `json:"lt,omitempty"`
	Lte string `json:"lte,omitempty"`
}

func (r Range) IsZero() bool {
	return r.Gt == "" && r.Gte == "" && r.Lt == "" && r.Lte == ""
}

type SortField struct {
	Field string `json:"field"`
	Desc  bool   `json:"desc"`
}

type DocumentQuery struct {
	Keyword string              `json:"q"`
	Indices []string            `json:"indices"`
	Filters map[string][]string `json:"filters"`
	Ranges  map[string]Range    `json:"ranges"`
	Sort    []SortField         `json:"sort"`
	Facets  []string            `json:"facets"`
	Size    int                 `json:"-"`
	Cursor  string              `json:"-"`
}

func (q *DocumentQuery) Normalize(mode QueryMode) error {
	q.Keyword = strings.TrimSpace(q.Keyword)
	if mode == ModeSearch && q.Keyword == "" {
		return Invalid("q is required")
	}
	if len(q.Keyword) > MaxKeywordLen {
		return Invalid("q must be at most %d characters", MaxKeywordLen)
	}

	indices, err := normalizeIndices(q.Indices)
	if err != nil {
		return err
	}
	q.Indices = indices

	if q.Size, err = normalizeSize(q.Size, DefaultPageSize, MaxPageSize); err != nil {
		return err
	}

	if len(q.Filters)+len(q.Ranges) > MaxFilters {
		return Invalid("at most %d filters are allowed", MaxFilters)
	}
	filters := make(map[string][]string, len(q.Filters))
	for field, values := range q.Filters {
		if err := ValidateField(field); err != nil {
			return err
		}
		if values = SplitList(values); len(values) > 0 {
			filters[field] = values
		}
	}
	q.Filters = filters

	ranges := make(map[string]Range, len(q.Ranges))
	for field, r := range q.Ranges {
		if err := ValidateField(field); err != nil {
			return err
		}
		if !r.IsZero() {
			ranges[field] = r
		}
	}
	q.Ranges = ranges

	if len(q.Sort) > MaxSortFields {
		return Invalid("at most %d sort fields are allowed", MaxSortFields)
	}
	for _, s := range q.Sort {
		if s.Field == ScoreField {
			continue
		}
		if err := ValidateField(s.Field); err != nil {
			return err
		}
	}

	q.Facets = SplitList(q.Facets)
	if len(q.Facets) > MaxFacets {
		return Invalid("at most %d facets are allowed", MaxFacets)
	}
	for _, f := range q.Facets {
		if f == IndexFacet {
			continue
		}
		if err := ValidateField(f); err != nil {
			return err
		}
	}
	return nil
}

func (q DocumentQuery) Fields() []string {
	var fields []string
	for f := range q.Filters {
		fields = append(fields, f)
	}
	for f := range q.Ranges {
		fields = append(fields, f)
	}
	for _, s := range q.Sort {
		if s.Field != ScoreField {
			fields = append(fields, s.Field)
		}
	}
	for _, f := range q.Facets {
		if f != IndexFacet {
			fields = append(fields, f)
		}
	}
	return fields
}

func (q DocumentQuery) Fingerprint() string {
	data, _ := json.Marshal(q)
	return string(data)
}

type AutocompleteQuery struct {
	Keyword string
	Indices []string
	Size    int
}

func (q *AutocompleteQuery) Normalize() error {
	q.Keyword = strings.TrimSpace(q.Keyword)
	if q.Keyword == "" {
		return Invalid("q is required")
	}
	if len(q.Keyword) > MaxKeywordLen {
		return Invalid("q must be at most %d characters", MaxKeywordLen)
	}
	indices, err := normalizeIndices(q.Indices)
	if err != nil {
		return err
	}
	q.Indices = indices
	q.Size, err = normalizeSize(q.Size, DefaultAutocompleteSize, MaxAutocompleteSize)
	return err
}

type DocumentRef struct {
	Index string
	ID    string
}

func (r DocumentRef) Validate() error {
	if err := ValidateIndex(r.Index, false); err != nil {
		return err
	}
	if r.ID == "" || len(r.ID) > 512 {
		return Invalid("invalid document id")
	}
	return nil
}

func NormalizeSimilarSize(size int) (int, error) {
	return normalizeSize(size, DefaultSimilarSize, MaxPageSize)
}

func ParseSort(raw string) []SortField {
	var out []SortField
	for _, part := range SplitList([]string{raw}) {
		field, dir, _ := strings.Cut(part, ":")
		out = append(out, SortField{Field: strings.TrimSpace(field), Desc: strings.EqualFold(strings.TrimSpace(dir), "desc")})
	}
	return out
}

func ValidateField(field string) error {
	if !fieldNamePattern.MatchString(field) {
		return Invalid("invalid field %q", field)
	}
	return nil
}

func ValidateIndex(name string, allowWildcard bool) error {
	if !indexNamePattern.MatchString(name) || (!allowWildcard && strings.Contains(name, "*")) {
		return Invalid("invalid index %q", name)
	}
	return nil
}

func normalizeIndices(indices []string) ([]string, error) {
	out := SplitList(indices)
	for _, name := range out {
		if err := ValidateIndex(name, true); err != nil {
			return nil, err
		}
	}
	return out, nil
}

func normalizeSize(size, def, max int) (int, error) {
	if size < 0 || size > max {
		return 0, Invalid("size must be between 1 and %d", max)
	}
	if size == 0 {
		return def, nil
	}
	return size, nil
}

func SplitList(values []string) []string {
	out := make([]string, 0, len(values))
	for _, raw := range values {
		for _, v := range strings.Split(raw, ",") {
			if v = strings.TrimSpace(v); v != "" {
				out = append(out, v)
			}
		}
	}
	return out
}
