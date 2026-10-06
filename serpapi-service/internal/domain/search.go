package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"regexp"
	"sort"
	"strings"
)

type Params map[string]string

var reservedParams = map[string]bool{
	"api_key":  true,
	"engine":   true,
	"output":   true,
	"async":    true,
	"no_cache": true,
}

func IsReservedParam(name string) bool { return reservedParams[name] }

type Output string

const (
	OutputJSON          Output = "json"
	OutputHTML          Output = "html"
	OutputMarkdown      Output = "md"
	OutputPixelPosition Output = "json_with_pixel_position"
)

var pixelPositionEngines = map[string]bool{"google": true, "google_ads": true}

func SupportsPixelPosition(engine string) bool { return pixelPositionEngines[engine] }

func (o Output) IsJSON() bool { return o == OutputJSON || o == OutputPixelPosition }

func ParseOutput(s string) (Output, error) {
	switch strings.ToLower(s) {
	case "", "json":
		return OutputJSON, nil
	case "html":
		return OutputHTML, nil
	case "md":
		return OutputMarkdown, nil
	case "json_with_pixel_position":
		return OutputPixelPosition, nil
	}
	return "", Invalid("output must be json, html, md or json_with_pixel_position, got %q", s)
}

type SearchRequest struct {
	Engine  string
	Params  Params
	Output  Output
	NoCache bool
	Async   bool
}

func (r SearchRequest) ZeroTrace() bool {
	v := strings.ToLower(r.Params["zero_trace"])
	return v == "true" || v == "1"
}

func (r SearchRequest) CacheKey() string {
	keys := make([]string, 0, len(r.Params))
	for k := range r.Params {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	h := sha256.New()
	h.Write([]byte(r.Engine))
	h.Write([]byte{0})
	h.Write([]byte(r.Output))
	for _, k := range keys {
		h.Write([]byte{0})
		h.Write([]byte(k))
		h.Write([]byte{'='})
		h.Write([]byte(r.Params[k]))
	}
	return "search:" + hex.EncodeToString(h.Sum(nil))
}

type SearchResult struct {
	Body        []byte `json:"body"`
	ContentType string `json:"content_type"`
	SearchID    string `json:"search_id,omitempty"`
	Status      string `json:"status,omitempty"`
	Cached      bool   `json:"-"`
}

const (
	StatusSuccess = "Success"
)

var searchIDPattern = regexp.MustCompile(`^[A-Za-z0-9_-]{1,64}$`)

func ValidateSearchID(id string) error {
	if !searchIDPattern.MatchString(id) {
		return Invalid("invalid search id %q", id)
	}
	return nil
}
