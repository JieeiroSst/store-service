package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"

	"github.com/JIeeiroSst/search-service/config"
	"github.com/JIeeiroSst/search-service/internal/domain"
	esv7 "github.com/elastic/go-elasticsearch/v7"
	esv7api "github.com/elastic/go-elasticsearch/v7/esapi"
)

var (
	defaultIndices = []string{"*", "-.*"}
	versionSuffix  = regexp.MustCompile(`__v[0-9]+$`)
)

func NewClient(cfg *config.Config) (*esv7.Client, error) {
	client, err := esv7.NewClient(esv7.Config{
		Addresses: []string{cfg.Elasticsearch.DNS},
		Username:  cfg.Elasticsearch.Username,
		Password:  cfg.Elasticsearch.Password,
	})
	if err != nil {
		return nil, fmt.Errorf("create elasticsearch client: %w", err)
	}
	return client, nil
}

type Health struct {
	client *esv7.Client
}

func NewHealth(client *esv7.Client) *Health {
	return &Health{client: client}
}

func (h *Health) Ping(ctx context.Context) error {
	return call(ctx, h.client, esv7api.PingRequest{}, nil)
}

func call(ctx context.Context, client *esv7.Client, req esv7api.Request, out any) error {
	resp, err := req.Do(ctx, client)
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return err
		}
		return fmt.Errorf("%w: %v", domain.ErrUnavailable, err)
	}
	defer resp.Body.Close()

	if resp.IsError() {
		return responseError(resp)
	}
	if out == nil {
		_, _ = io.Copy(io.Discard, resp.Body)
		return nil
	}
	dec := json.NewDecoder(resp.Body)
	dec.UseNumber()
	if err := dec.Decode(out); err != nil {
		return fmt.Errorf("decode elasticsearch response: %w", err)
	}
	return nil
}

func responseError(resp *esv7api.Response) error {
	var body struct {
		Error struct {
			Type      string `json:"type"`
			Reason    string `json:"reason"`
			RootCause []struct {
				Reason string `json:"reason"`
			} `json:"root_cause"`
		} `json:"error"`
	}
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 64<<10))
	_ = json.Unmarshal(raw, &body)

	reason := body.Error.Reason
	if len(body.Error.RootCause) > 0 && body.Error.RootCause[0].Reason != "" {
		reason = body.Error.RootCause[0].Reason
	}

	switch {
	case resp.StatusCode == http.StatusNotFound:
		if reason == "" {
			return domain.ErrNotFound
		}
		return fmt.Errorf("%w: %s", domain.ErrNotFound, reason)
	case resp.StatusCode == http.StatusBadRequest:
		return domain.Invalid("%s", reason)
	case resp.StatusCode == http.StatusServiceUnavailable || resp.StatusCode == http.StatusTooManyRequests:
		return fmt.Errorf("%w: %s", domain.ErrUnavailable, reason)
	}
	if reason == "" {
		reason = strings.TrimSpace(string(raw))
	}
	return fmt.Errorf("elasticsearch %d %s: %s", resp.StatusCode, body.Error.Type, reason)
}

func jsonBody(v any) (io.Reader, error) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(v); err != nil {
		return nil, fmt.Errorf("encode elasticsearch request: %w", err)
	}
	return &buf, nil
}

func displayIndex(index string) string {
	return versionSuffix.ReplaceAllString(index, "")
}

func boolPtr(v bool) *bool {
	return &v
}
