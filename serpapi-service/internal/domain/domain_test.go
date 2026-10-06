package domain

import "testing"

func TestCatalogHasUniqueValidEngines(t *testing.T) {
	seen := map[string]bool{}
	for _, e := range Engines {
		if !ValidEngineID(e.ID) {
			t.Errorf("invalid engine id %q", e.ID)
		}
		if seen[e.ID] {
			t.Errorf("duplicate engine %q", e.ID)
		}
		seen[e.ID] = true
		if e.Group == "" || e.Name == "" {
			t.Errorf("engine %q needs a group and a name", e.ID)
		}
	}
	c := NewCatalog(Engines)
	if len(c.List("")) != len(Engines) {
		t.Fatalf("catalog lost engines")
	}
	if got := c.List("youtube"); len(got) != 4 {
		t.Errorf("youtube group: got %d engines", len(got))
	}
}

func TestEngineValidate(t *testing.T) {
	c := NewCatalog(Engines)
	cases := []struct {
		engine string
		params Params
		ok     bool
	}{
		{"google", Params{"q": "coffee"}, true},
		{"google", Params{"q": "  "}, false},
		{"google", Params{}, false},
		{"google_maps_reviews", Params{"place_id": "x"}, true},
		{"google_maps_reviews", Params{"data_id": "x"}, true},
		{"walmart_product_sellers", Params{"product_id": "1"}, false},
		{"apple_maps", Params{"query": "cafe", "location": "x"}, true},
		{"apple_maps", Params{"query": "cafe"}, false},
		{"google_news", Params{}, true},
		{"google_local_services", Params{"q": "plumber", "data_cid": "1"}, true},
		{"google_local_services", Params{"cid": "1", "bid": "2", "pid": "3"}, true},
		{"google_local_services", Params{}, false},
		{"google_scholar_profiles", Params{"mauthors": "x"}, false},
		{"google_lens_image_sources", Params{"page_token": "x"}, false},
	}
	for _, tc := range cases {
		e, ok := c.Get(tc.engine)
		if !ok {
			t.Fatalf("missing engine %s", tc.engine)
		}
		if err := e.Validate(tc.params); (err == nil) != tc.ok {
			t.Errorf("%s %v: err=%v, want ok=%v", tc.engine, tc.params, err, tc.ok)
		} else if err != nil && !IsInvalid(err) {
			t.Errorf("want InvalidError, got %T", err)
		}
	}
}

func TestCacheKeyIgnoresParamOrder(t *testing.T) {
	a := SearchRequest{Engine: "google", Output: OutputJSON, Params: Params{"q": "x", "hl": "en"}}
	b := SearchRequest{Engine: "google", Output: OutputJSON, Params: Params{"hl": "en", "q": "x"}}
	if a.CacheKey() != b.CacheKey() {
		t.Fatal("same request, different keys")
	}
	b.Output = OutputHTML
	if a.CacheKey() == b.CacheKey() {
		t.Fatal("output must be part of the key")
	}

	c := SearchRequest{Engine: "google", Params: Params{"q": "a=b"}}
	d := SearchRequest{Engine: "google", Params: Params{"q": "a", "b": ""}}
	if c.CacheKey() == d.CacheKey() {
		t.Fatal("collision")
	}
}

func TestParseOutputAndIDs(t *testing.T) {
	for in, want := range map[string]Output{"": OutputJSON, "JSON": OutputJSON, "html": OutputHTML, "md": OutputMarkdown, "json_with_pixel_position": OutputPixelPosition} {
		if got, err := ParseOutput(in); err != nil || got != want {
			t.Errorf("ParseOutput(%q) = %q, %v", in, got, err)
		}
	}
	if _, err := ParseOutput("xml"); !IsInvalid(err) {
		t.Error("xml must be invalid")
	}
	if ValidateSearchID("5b50d58a304bda2fca30bac9") != nil {
		t.Error("valid id rejected")
	}
	for _, id := range []string{"", "../account", "a/b", "a.json"} {
		if ValidateSearchID(id) == nil {
			t.Errorf("id %q accepted", id)
		}
	}
}

func TestNewImageUploadSniffsType(t *testing.T) {
	cases := map[string]string{
		"\xFF\xD8\xFF\xE0":            "image/jpeg",
		"\x89PNG\r\n\x1a\n....":       "image/png",
		"RIFF\x00\x00\x00\x00WEBPVP8": "image/webp",
	}
	for data, want := range cases {
		img, err := NewImageUpload("", []byte(data))
		if err != nil || img.ContentType != want || img.Filename == "" {
			t.Errorf("%q: %+v %v", data, img, err)
		}
	}
	for _, bad := range []string{"", "GIF89a", "<svg/>"} {
		if _, err := NewImageUpload("x", []byte(bad)); !IsInvalid(err) {
			t.Errorf("%q accepted", bad)
		}
	}
	if !OutputPixelPosition.IsJSON() || OutputHTML.IsJSON() || !SupportsPixelPosition("google_ads") || SupportsPixelPosition("bing") {
		t.Error("output helpers")
	}
}
