package browserqa

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/png"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

const goodPage = `<!doctype html><html lang="en"><head><meta charset="utf-8"><title>Shop</title><meta name="viewport" content="width=device-width,initial-scale=1">
<style>body{font:16px sans-serif;color:#111;background:#fff;margin:0;padding:8px}a,button,input{min-width:44px;min-height:44px;display:inline-block;box-sizing:border-box}</style></head>
<body><main><h1>Shop</h1><p>Welcome to our store.</p><a href="/about">About</a> <button>Buy now</button>
<label for="q">Search</label><input id="q" type="text"></main></body></html>`

const badPage = `<!doctype html><html><head><meta charset="utf-8"><title>Bad</title>
<style>body{font:16px sans-serif;background:#fff;margin:0}
.grey{color:#bbb}
.wide{width:1200px;height:20px;background:#eee}
button{outline:none;border:1px solid #999}button:focus{box-shadow:none}
.tiny{display:inline-block;width:10px;height:10px;background:#333}</style></head>
<body><div class="grey">low contrast text here</div><div class="wide"></div>
<button></button> <a href="/x"><img src="data:image/gif;base64,R0lGODlhAQABAAAAACH5BAEKAAEALAAAAAABAAEAAAICTAEAOw=="></a>
<div onclick="go()" style="cursor:pointer">clickable div</div>
<a class="tiny" href="/t"></a><p>Hello {{username}}!</p>
<script>console.error("boom from console");setTimeout(function(){throw new Error("uncaught failure")},50)</script></body></html>`

const trapPage = `<!doctype html><html lang="en"><head><title>Trap</title></head><body><main><h1>Trap</h1>
<input id="a" aria-label="a" onblur="setTimeout(function(){document.getElementById('a').focus()},0)"><button>after</button></main></body></html>`

func hasRule(r Result, rule string) bool {
	for _, i := range r.Issues {
		if i.Rule == rule {
			return true
		}
	}
	return false
}

func server(pages map[string]string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if p, ok := pages[r.URL.Path]; ok {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Write([]byte(p))
			return
		}
		http.NotFound(w, r)
	}))
}

func needChrome(t *testing.T) {
	if _, err := FindChrome(); err != nil {
		t.Skip("no Chrome/Chromium available")
	}
}

func opts(t *testing.T) Options {
	return Options{NoSandbox: true, ArtifactDir: t.TempDir(), BaselineDir: t.TempDir(), Name: "t"}
}

func TestDefectivePageIsFlagged(t *testing.T) {
	needChrome(t)
	srv := server(map[string]string{"/bad": badPage})
	defer srv.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	rep, err := Audit(ctx, []string{srv.URL + "/bad"}, opts(t))
	if err != nil {
		t.Fatal(err)
	}
	all := map[string]bool{}
	byVP := map[string]Result{}
	for _, r := range rep.Results {
		byVP[r.Viewport] = r
		for _, i := range r.Issues {
			all[r.Viewport+":"+i.Rule] = true
		}
	}
	for _, want := range []string{
		"desktop:contrast", "desktop:js-exception", "desktop:console-error", "desktop:i18n-leak",
		"desktop:unnamed-button", "desktop:click-not-keyboard", "desktop:no-focus-indicator", "desktop:no-main-landmark",
		"mobile:horizontal-overflow", "mobile:tap-target", "mobile:no-viewport-meta",
	} {
		if !all[want] {
			t.Errorf("missing %s; got %v", want, all)
		}
	}
	if all["desktop:tap-target"] {
		t.Error("tap targets are only checked on touch viewports")
	}
	if r := byVP["desktop"]; r.Screenshot == "" || r.Baseline != "created" {
		t.Errorf("screenshot/baseline: %+v", r)
	}
}

func TestGoodPageIsClean(t *testing.T) {
	needChrome(t)
	srv := server(map[string]string{"/ok": goodPage})
	defer srv.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	rep, err := Audit(ctx, []string{srv.URL + "/ok"}, opts(t))
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range rep.Results {
		for _, i := range r.Issues {
			if i.Severity != Note {
				t.Errorf("[%s] well-formed page flagged: %+v", r.Viewport, i)
			}
		}
	}
	if rep.Results[0].TabStops < 3 {
		t.Errorf("expected keyboard to reach link, button and input; got %d stops", rep.Results[0].TabStops)
	}
}

func TestKeyboardTrap(t *testing.T) {
	needChrome(t)
	srv := server(map[string]string{"/trap": trapPage})
	defer srv.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	o := opts(t)
	o.Viewports = DefaultViewports[:1]
	rep, err := Audit(ctx, []string{srv.URL + "/trap"}, o)
	if err != nil {
		t.Fatal(err)
	}
	if !hasRule(rep.Results[0], "keyboard-trap") {
		t.Fatalf("trap not detected: %+v", rep.Results[0].Issues)
	}
}

func TestLocalization(t *testing.T) {
	needChrome(t)
	srv := server(map[string]string{"/ok": goodPage})
	defer srv.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	o := opts(t)
	o.Viewports = DefaultViewports[:1]
	o.Locales = []string{"en-US", "ar"}
	rep, err := Audit(ctx, []string{srv.URL + "/ok"}, o)
	if err != nil {
		t.Fatal(err)
	}
	var rtl, notLoc bool
	for _, r := range rep.Results {
		rtl = rtl || hasRule(r, "rtl-not-applied")
		notLoc = notLoc || hasRule(r, "not-localized")
	}
	if !rtl || !notLoc {
		t.Fatalf("rtl=%v notLocalized=%v results=%+v", rtl, notLoc, rep.Results)
	}
}

func TestVisualRegression(t *testing.T) {
	needChrome(t)
	page := goodPage
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(page))
	}))
	defer srv.Close()
	o := opts(t)
	o.Viewports = DefaultViewports[:1]
	run := func() Result {
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()
		rep, err := Audit(ctx, []string{srv.URL}, o)
		if err != nil {
			t.Fatal(err)
		}
		return rep.Results[0]
	}
	if r := run(); r.Baseline != "created" {
		t.Fatalf("first run must create the baseline: %+v", r)
	}
	if r := run(); r.Baseline != "compared" || r.DiffPct == nil || *r.DiffPct > 0.5 || hasRule(r, "visual-regression") {
		t.Fatalf("unchanged page must match its baseline: %+v", r)
	}
	page = strings.Replace(goodPage, "background:#fff", "background:#ffeeee", 1)
	page = strings.Replace(page, "<p>Welcome to our store.</p>", "<p>Welcome to our store.</p><div style='height:200px;background:#123456'></div>", 1)
	if r := run(); !hasRule(r, "visual-regression") || r.DiffImage == "" {
		t.Fatalf("visual change must be flagged: %+v", r)
	}
	o.UpdateBaseline = true
	if r := run(); r.Baseline != "updated" {
		t.Fatalf("update must refresh the baseline: %+v", r)
	}
}

func TestAllowListBlocksExternalRequests(t *testing.T) {
	needChrome(t)
	external := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		os.Stderr.WriteString("")
		http.Error(w, "reached", 200)
	}))
	defer external.Close()
	hit := make(chan struct{}, 1)
	ext := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { hit <- struct{}{} }))
	defer ext.Close()

	page := `<html lang="en"><head><title>x</title></head><body><main><h1>x</h1><img alt="" src="http://localhost:` + strings.Split(ext.URL, ":")[2] + `/pixel.gif"></main></body></html>`
	srv := server(map[string]string{"/p": page})
	defer srv.Close()

	o := opts(t)
	o.Viewports = DefaultViewports[:1]
	o.AllowedHosts = map[string]bool{"127.0.0.1": true}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	if _, err := Audit(ctx, []string{srv.URL + "/p"}, o); err != nil {
		t.Fatal(err)
	}
	select {
	case <-hit:
		t.Fatal("browser reached a host outside the allow-list")
	default:
	}
}

func pngOf(c color.Color, w, h int, mark bool) []byte {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, c)
		}
	}
	if mark {
		for y := 0; y < 10; y++ {
			for x := 0; x < 10; x++ {
				img.Set(x, y, color.RGBA{0, 0, 0, 255})
			}
		}
	}
	var b bytes.Buffer
	png.Encode(&b, img)
	return b.Bytes()
}

func TestDiffImages(t *testing.T) {
	white := color.RGBA{255, 255, 255, 255}
	pct, same, diff, err := diffImages(pngOf(white, 100, 100, false), pngOf(white, 100, 100, false))
	if err != nil || !same || pct != 0 || diff == nil {
		t.Fatalf("identical: %v %v %v", pct, same, err)
	}
	pct, _, _, _ = diffImages(pngOf(white, 100, 100, false), pngOf(white, 100, 100, true))
	if pct < 0.99 || pct > 1.01 {
		t.Fatalf("expected 1%%, got %v", pct)
	}

	pct, _, _, _ = diffImages(pngOf(color.RGBA{250, 250, 250, 255}, 50, 50, false), pngOf(color.RGBA{255, 255, 255, 255}, 50, 50, false))
	if pct != 0 {
		t.Fatalf("noise must be tolerated, got %v", pct)
	}
	if pct, same, _, _ := diffImages(pngOf(white, 100, 100, false), pngOf(white, 100, 120, false)); same || pct != 100 {
		t.Fatalf("size change: %v %v", pct, same)
	}
	if _, _, _, err := diffImages([]byte("junk"), pngOf(white, 5, 5, false)); err == nil {
		t.Fatal("corrupt baseline must error")
	}
}
