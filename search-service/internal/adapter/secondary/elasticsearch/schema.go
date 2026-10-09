package elasticsearch

import (
	"context"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/JIeeiroSst/search-service/config"
	"github.com/JIeeiroSst/search-service/internal/domain"
	esv7 "github.com/elastic/go-elasticsearch/v7"
	esv7api "github.com/elastic/go-elasticsearch/v7/esapi"
)

const (
	SchemaVersion        = 1
	defaultTemplateName  = "store-sync"
	defaultPriority      = 10
	sensitivePlaceholder = "__SENSITIVE__"
	rollbackTimeout      = time.Minute
)

//go:embed schema/*.json
var schemaFiles embed.FS

type SchemaManager struct {
	client   *esv7.Client
	searcher *Searcher
	cfg      config.SchemaConfig
	policy   domain.SensitivePolicy
	now      func() time.Time
}

func NewSchemaManager(client *esv7.Client, searcher *Searcher, cfg *config.Config, policy domain.SensitivePolicy) *SchemaManager {
	s := cfg.Elasticsearch.Schema
	if s.Name == "" {
		s.Name = defaultTemplateName
	}
	if len(s.IndexPatterns) == 0 {
		s.IndexPatterns = []string{"*"}
	}
	if s.Priority == 0 {
		s.Priority = defaultPriority
	}
	return &SchemaManager{client: client, searcher: searcher, cfg: s, policy: policy, now: time.Now}
}

func (s *SchemaManager) Apply(ctx context.Context) error {
	settings, err := s.settings()
	if err != nil {
		return err
	}
	mappings, err := s.mappings()
	if err != nil {
		return err
	}
	meta := map[string]any{"managed_by": "search-service", "version": SchemaVersion}

	component := s.cfg.Name + "-schema"
	body, err := jsonBody(map[string]any{
		"template": map[string]any{"settings": settings, "mappings": mappings},
		"version":  SchemaVersion,
		"_meta":    meta,
	})
	if err != nil {
		return err
	}
	if err := call(ctx, s.client, esv7api.ClusterPutComponentTemplateRequest{Name: component, Body: body}, nil); err != nil {
		return fmt.Errorf("put component template %s: %w", component, err)
	}

	body, err = jsonBody(map[string]any{
		"index_patterns": s.cfg.IndexPatterns,
		"priority":       s.cfg.Priority,
		"composed_of":    []string{component},
		"version":        SchemaVersion,
		"_meta":          meta,
	})
	if err != nil {
		return err
	}
	if err := call(ctx, s.client, esv7api.IndicesPutIndexTemplateRequest{Name: s.cfg.Name, Body: body}, nil); err != nil {
		return fmt.Errorf("put index template %s: %w", s.cfg.Name, err)
	}
	return nil
}

func (s *SchemaManager) Status(ctx context.Context) (domain.SchemaStatus, error) {
	status := domain.SchemaStatus{Template: s.cfg.Name}

	var res struct {
		IndexTemplates []struct {
			IndexTemplate struct {
				Version int `json:"version"`
			} `json:"index_template"`
		} `json:"index_templates"`
	}
	err := call(ctx, s.client, esv7api.IndicesGetIndexTemplateRequest{Name: s.cfg.Name}, &res)
	switch {
	case err == nil && len(res.IndexTemplates) > 0:
		status.Applied = true
		status.Version = res.IndexTemplates[0].IndexTemplate.Version
	case err != nil && !errors.Is(err, domain.ErrNotFound):
		return domain.SchemaStatus{}, err
	}

	indices, err := s.searcher.Indices(ctx)
	if err != nil {
		return domain.SchemaStatus{}, err
	}
	status.Indices = indices
	return status, nil
}

func (s *SchemaManager) Reindex(ctx context.Context, name string) (domain.ReindexResult, error) {
	status, err := s.Status(ctx)
	if err != nil {
		return domain.ReindexResult{}, err
	}
	if !status.Applied || status.Version != SchemaVersion {
		return domain.ReindexResult{}, domain.Invalid("schema template %s v%d is not applied", s.cfg.Name, SchemaVersion)
	}

	var existing map[string]struct {
		Aliases map[string]any `json:"aliases"`
	}
	if err := call(ctx, s.client, esv7api.IndicesGetRequest{Index: []string{name}}, &existing); err != nil {
		return domain.ReindexResult{}, err
	}
	if len(existing) != 1 {
		return domain.ReindexResult{}, domain.Invalid("%q resolves to %d indices, expected 1", name, len(existing))
	}

	var oldIndex string
	aliases := map[string]bool{}
	for index, info := range existing {
		oldIndex = index
		for alias := range info.Aliases {
			aliases[alias] = true
		}
	}
	baseName := displayIndex(oldIndex)
	if name != oldIndex {
		baseName = name
	}
	aliases[baseName] = true
	newIndex := baseName + "__v" + strconv.FormatInt(s.now().Unix(), 10)
	result := domain.ReindexResult{Name: baseName, From: oldIndex, To: newIndex}

	if err := s.setWriteBlock(ctx, oldIndex, true); err != nil {
		return result, err
	}
	rollback := func(cause error) (domain.ReindexResult, error) {
		rctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), rollbackTimeout)
		defer cancel()
		_ = call(rctx, s.client, esv7api.IndicesDeleteRequest{Index: []string{newIndex}}, nil)
		_ = s.setWriteBlock(rctx, oldIndex, false)
		return result, cause
	}

	if err := call(ctx, s.client, esv7api.IndicesCreateRequest{Index: newIndex}, nil); err != nil {
		return rollback(err)
	}

	body, err := jsonBody(map[string]any{
		"source": map[string]any{"index": oldIndex},
		"dest":   map[string]any{"index": newIndex},
	})
	if err != nil {
		return rollback(err)
	}
	var reindexed struct {
		Created  int64 `json:"created"`
		Updated  int64 `json:"updated"`
		Failures []any `json:"failures"`
	}
	if err := call(ctx, s.client, esv7api.ReindexRequest{Body: body, WaitForCompletion: boolPtr(true), Refresh: boolPtr(true)}, &reindexed); err != nil {
		return rollback(err)
	}
	if len(reindexed.Failures) > 0 {
		return rollback(fmt.Errorf("reindex %s had %d failures: %v", oldIndex, len(reindexed.Failures), reindexed.Failures[0]))
	}
	result.Copied = reindexed.Created + reindexed.Updated

	actions := []any{map[string]any{"remove_index": map[string]any{"index": oldIndex}}}
	for _, alias := range sortedKeys(aliases) {
		actions = append(actions, map[string]any{"add": map[string]any{"index": newIndex, "alias": alias}})
	}
	aliasBody, err := jsonBody(map[string]any{"actions": actions})
	if err != nil {
		return rollback(err)
	}
	if err := call(ctx, s.client, esv7api.IndicesUpdateAliasesRequest{Body: aliasBody}, nil); err != nil {
		return rollback(err)
	}
	return result, nil
}

func (s *SchemaManager) setWriteBlock(ctx context.Context, index string, blocked bool) error {
	var value any
	if blocked {
		value = true
	}
	body, err := jsonBody(map[string]any{"index.blocks.write": value})
	if err != nil {
		return err
	}
	return call(ctx, s.client, esv7api.IndicesPutSettingsRequest{Index: []string{index}, Body: body}, nil)
}

func (s *SchemaManager) settings() (map[string]any, error) {
	var settings map[string]any
	if err := readSchemaFile("schema/settings.json", &settings); err != nil {
		return nil, err
	}
	index := settings["index"].(map[string]any)
	if s.cfg.NumberOfShards > 0 {
		index["number_of_shards"] = s.cfg.NumberOfShards
	}
	if s.cfg.NumberOfReplicas != nil {
		index["number_of_replicas"] = *s.cfg.NumberOfReplicas
	}
	if len(s.cfg.Synonyms) > 0 {
		analysis := settings["analysis"].(map[string]any)
		analysis["filter"].(map[string]any)["synonyms"] = map[string]any{
			"type":     "synonym_graph",
			"synonyms": s.cfg.Synonyms,
			"lenient":  true,
		}
		analyzer := analysis["analyzer"].(map[string]any)["text_search"].(map[string]any)
		analyzer["filter"] = []string{"lowercase", "synonyms", "folding"}
	}
	return settings, nil
}

func (s *SchemaManager) mappings() (map[string]any, error) {
	var mappings map[string]any
	if err := readSchemaFile("schema/mappings.json", &mappings); err != nil {
		return nil, err
	}
	for _, entry := range mappings["dynamic_templates"].([]any) {
		for _, tpl := range entry.(map[string]any) {
			t := tpl.(map[string]any)
			if t["match"] == sensitivePlaceholder {
				t["match"] = s.policy.Pattern()
			}
		}
	}
	return mappings, nil
}

func readSchemaFile(name string, out any) error {
	data, err := schemaFiles.ReadFile(name)
	if err != nil {
		return fmt.Errorf("read %s: %w", name, err)
	}
	if err := json.Unmarshal(data, out); err != nil {
		return fmt.Errorf("parse %s: %w", name, err)
	}
	return nil
}
