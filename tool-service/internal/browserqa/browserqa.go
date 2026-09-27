package browserqa

import (
	"context"
	"crypto/sha1"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/chromedp/cdproto/accessibility"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/cdproto/fetch"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
	"github.com/chromedp/chromedp/kb"
)

const (
	Error = "error"
	Warn  = "warn"
	Note  = "info"
)

type Viewport struct {
	Name   string  `json:"name"`
	W      int64   `json:"width"`
	H      int64   `json:"height"`
	Mobile bool    `json:"mobile"`
	Scale  float64 `json:"scale,omitempty"`
}

var DefaultViewports = []Viewport{
	{"desktop", 1366, 768, false, 1},
	{"tablet", 820, 1180, true, 2},
	{"mobile", 390, 844, true, 3},
}

type Options struct {
	Viewports        []Viewport
	Locales          []string
	Name             string
	BaselineDir      string
	ArtifactDir      string
	UpdateBaseline   bool
	DiffThresholdPct float64
	AllowedHosts     map[string]bool
	NoSandbox        bool
	Progress         func(string)
}

type Issue struct {
	Severity string `json:"severity"`
	Rule     string `json:"rule"`
	Detail   string `json:"detail"`
}

type Result struct {
	Page       string   `json:"page"`
	Viewport   string   `json:"viewport"`
	Locale     string   `json:"locale,omitempty"`
	Issues     []Issue  `json:"issues"`
	Screenshot string   `json:"screenshot,omitempty"`
	Baseline   string   `json:"baseline,omitempty"`
	DiffPct    *float64 `json:"visual_diff_pct,omitempty"`
	DiffImage  string   `json:"diff_image,omitempty"`
	TabStops   int      `json:"tab_stops,omitempty"`
}

type Report struct {
	Browser  string   `json:"browser"`
	Results  []Result `json:"results"`
	Errors   int      `json:"error_count"`
	Warnings int      `json:"warning_count"`
}

var ErrNoBrowser = errors.New("no Chrome/Chromium found: set CHROME_PATH or install chromium")

func FindChrome() (string, error) {
	if p := os.Getenv("CHROME_PATH"); p != "" {
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}
	for _, c := range []string{"chromium", "chromium-browser", "google-chrome", "google-chrome-stable", "chrome",
		"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome", "/Applications/Chromium.app/Contents/MacOS/Chromium"} {
		if p, err := exec.LookPath(c); err == nil {
			return p, nil
		}
		if strings.HasPrefix(c, "/") {
			if _, err := os.Stat(c); err == nil {
				return c, nil
			}
		}
	}
	return "", ErrNoBrowser
}

func Audit(ctx context.Context, urls []string, opt Options) (*Report, error) {
	chromePath, err := FindChrome()
	if err != nil {
		return nil, err
	}
	if len(opt.Viewports) == 0 {
		opt.Viewports = DefaultViewports
	}
	if len(opt.Locales) == 0 {
		opt.Locales = []string{""}
	}
	if opt.DiffThresholdPct <= 0 {
		opt.DiffThresholdPct = 0.5
	}
	if opt.Name == "" {
		opt.Name = "default"
	}
	for _, d := range []string{opt.BaselineDir, opt.ArtifactDir} {
		if d != "" {
			if err := os.MkdirAll(d, 0o755); err != nil {
				return nil, err
			}
		}
	}

	allocOpts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.ExecPath(chromePath),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("hide-scrollbars", true),
		chromedp.Flag("force-color-profile", "srgb"),
		chromedp.Flag("font-render-hinting", "none"),
	)
	if opt.NoSandbox {
		allocOpts = append(allocOpts, chromedp.NoSandbox)
	}

	for _, f := range strings.Fields(os.Getenv("CHROME_FLAGS")) {
		name, val, hasVal := strings.Cut(strings.TrimPrefix(f, "--"), "=")
		if hasVal {
			allocOpts = append(allocOpts, chromedp.Flag(name, val))
		} else {
			allocOpts = append(allocOpts, chromedp.Flag(name, true))
		}
	}
	allocOpts = append(allocOpts, chromedp.WSURLReadTimeout(45*time.Second))
	if os.Getenv("CHROME_DEBUG") == "true" {
		allocOpts = append(allocOpts, chromedp.CombinedOutput(os.Stderr))
	}
	allocCtx, cancelAlloc := chromedp.NewExecAllocator(ctx, allocOpts...)
	defer cancelAlloc()
	browserCtx, cancelBrowser := chromedp.NewContext(allocCtx)
	defer cancelBrowser()

	started := make(chan error, 1)
	go func() { started <- chromedp.Run(browserCtx) }()
	select {
	case err = <-started:
	case <-time.After(60 * time.Second):
		cancelBrowser()
		err = errors.New("timed out")
	}
	if err != nil {
		return nil, fmt.Errorf("cannot start browser (set CHROME_DEBUG=true to see its output): %w", err)
	}

	rep := &Report{Browser: chromePath}
	for _, pageURL := range urls {
		textByLocale := map[string]string{}
		for vi, vp := range opt.Viewports {
			for li, loc := range opt.Locales {
				if ctx.Err() != nil {
					return rep, ctx.Err()
				}
				if opt.Progress != nil {
					opt.Progress(fmt.Sprintf("%s [%s %s]", pageURL, vp.Name, loc))
				}
				first := vi == 0 && li == 0
				res, text := auditOne(browserCtx, pageURL, vp, loc, first, opt)
				if vi == 0 {
					textByLocale[loc] = text
				}
				rep.Results = append(rep.Results, res)
			}
		}
		if len(opt.Locales) > 1 {
			if issue := localizationIssue(textByLocale); issue != nil {
				for i := len(rep.Results) - 1; i >= 0; i-- {
					if rep.Results[i].Page == pageURL {
						rep.Results[i].Issues = append(rep.Results[i].Issues, *issue)
						break
					}
				}
			}
		}
	}
	for _, r := range rep.Results {
		for _, is := range r.Issues {
			switch is.Severity {
			case Error:
				rep.Errors++
			case Warn:
				rep.Warnings++
			}
		}
	}
	return rep, nil
}

func localizationIssue(byLocale map[string]string) *Issue {
	seen := map[[20]byte][]string{}
	for loc, t := range byLocale {
		h := sha1.Sum([]byte(strings.TrimSpace(t)))
		seen[h] = append(seen[h], loc)
	}
	if len(seen) == 1 && len(byLocale) > 1 {
		return &Issue{Note, "not-localized", "page text is identical for all requested locales; the site may ignore Accept-Language"}
	}
	return nil
}

func auditOne(browserCtx context.Context, pageURL string, vp Viewport, locale string, deep bool, opt Options) (Result, string) {
	res := Result{Page: pageURL, Viewport: vp.Name, Locale: locale, Issues: []Issue{}}
	add := func(sev, rule, detail string) { res.Issues = append(res.Issues, Issue{sev, rule, detail}) }

	tabCtx, cancelTab := chromedp.NewContext(browserCtx)
	defer cancelTab()
	tabCtx, cancelT := context.WithTimeout(tabCtx, 60*time.Second)
	defer cancelT()

	reqURL := map[network.RequestID]string{}
	origin := ""
	if u, err := url.Parse(pageURL); err == nil {
		origin = u.Host
	}
	seenIssue := map[string]bool{}
	once := func(sev, rule, detail string) {
		if k := rule + detail; !seenIssue[k] && len(res.Issues) < 60 {
			seenIssue[k] = true
			add(sev, rule, detail)
		}
	}
	chromedp.ListenTarget(tabCtx, func(ev any) {
		switch e := ev.(type) {
		case *runtime.EventExceptionThrown:
			msg := e.ExceptionDetails.Text
			if e.ExceptionDetails.Exception != nil && e.ExceptionDetails.Exception.Description != "" {
				msg = e.ExceptionDetails.Exception.Description
			}
			once(Error, "js-exception", firstLine(msg))
		case *runtime.EventConsoleAPICalled:
			if e.Type == runtime.APITypeError {
				var parts []string
				for _, a := range e.Args {
					if s := remoteText(a); s != "" {
						parts = append(parts, s)
					}
				}
				once(Warn, "console-error", firstLine(strings.Join(parts, " ")))
			}
		case *network.EventRequestWillBeSent:
			reqURL[e.RequestID] = e.Request.URL
		case *network.EventLoadingFailed:
			if !e.Canceled && e.ErrorText != "net::ERR_BLOCKED_BY_CLIENT" {
				once(Warn, "request-failed", reqURL[e.RequestID]+" "+e.ErrorText)
			}
		case *network.EventResponseReceived:
			if e.Response.Status >= 400 {

				if u, err := url.Parse(e.Response.URL); err == nil && u.Host == origin && !strings.HasSuffix(u.Path, "/favicon.ico") {
					once(Warn, "http-error", fmt.Sprintf("%d %s", e.Response.Status, e.Response.URL))
				}
			}
		case *fetch.EventRequestPaused:
			go func(ev *fetch.EventRequestPaused) {
				c := chromedp.FromContext(tabCtx)
				if c == nil || c.Target == nil {
					return
				}
				ctx := cdp.WithExecutor(tabCtx, c.Target)
				if hostAllowed(ev.Request.URL, opt.AllowedHosts) {
					_ = fetch.ContinueRequest(ev.RequestID).Do(ctx)
				} else {
					_ = fetch.FailRequest(ev.RequestID, network.ErrorReasonBlockedByClient).Do(ctx)
				}
			}(e)
		}
	})

	setup := []chromedp.Action{
		network.Enable(),
		emulation.SetDeviceMetricsOverride(vp.W, vp.H, orOne(vp.Scale), vp.Mobile),
		emulation.SetTouchEmulationEnabled(vp.Mobile),
	}
	if locale != "" {
		setup = append(setup, emulation.SetLocaleOverride().WithLocale(locale),
			network.SetExtraHTTPHeaders(network.Headers{"Accept-Language": locale}))
	}
	if len(opt.AllowedHosts) > 0 {
		setup = append(setup, fetch.Enable())
	}
	if err := chromedp.Run(tabCtx, setup...); err != nil {
		add(Error, "browser-setup", err.Error())
		return res, ""
	}

	if err := chromedp.Run(tabCtx, chromedp.Navigate(pageURL), chromedp.WaitReady("body"), chromedp.Sleep(600*time.Millisecond)); err != nil {
		add(Error, "navigation", firstLine(err.Error()))
		return res, ""
	}

	var raw string
	if err := chromedp.Run(tabCtx, chromedp.Evaluate(fmt.Sprintf("%s(%t)", pageAudit, vp.Mobile), &raw)); err != nil {
		add(Error, "audit-script", firstLine(err.Error()))
		return res, ""
	}
	var pa struct {
		Overflow, Contrast, Taps, Leaks, ClickNoKbd []string
		Text, Lang, Dir                             string
		PosTab, InnerWidth                          int
	}
	if err := json.Unmarshal([]byte(raw), &pa); err != nil {
		add(Error, "audit-script", err.Error())
		return res, ""
	}
	if vp.Mobile && int64(pa.InnerWidth) > vp.W+1 {
		add(Warn, "no-viewport-meta", fmt.Sprintf("on a %dpx-wide device the page lays out at %dpx; add <meta name=\"viewport\" content=\"width=device-width, initial-scale=1\">", vp.W, pa.InnerWidth))
	}
	for _, o := range pa.Overflow {
		add(Warn, "horizontal-overflow", o)
	}
	for _, c := range pa.Contrast {
		add(Warn, "contrast", c)
	}
	for _, t := range pa.Taps {
		add(Warn, "tap-target", t)
	}
	for _, l := range pa.Leaks {
		add(Warn, "i18n-leak", fmt.Sprintf("text contains %q", l))
	}
	if locale != "" {
		lang := strings.ToLower(strings.SplitN(locale, "-", 2)[0])
		if pa.Lang != "" && !strings.HasPrefix(strings.ToLower(pa.Lang), lang) {
			add(Note, "lang-mismatch", fmt.Sprintf("requested %s but <html lang=%q>", locale, pa.Lang))
		}
		if map[string]bool{"ar": true, "he": true, "fa": true, "ur": true}[lang] && pa.Dir != "rtl" {
			add(Warn, "rtl-not-applied", "locale "+locale+" is right-to-left but the page direction is "+pa.Dir)
		}
	}

	if deep {
		res.TabStops = keyboardAudit(tabCtx, add, pa.ClickNoKbd, pa.PosTab)
		axAudit(tabCtx, add)
	}

	var shot []byte
	if err := chromedp.Run(tabCtx, chromedp.FullScreenshot(&shot, 100)); err != nil {
		add(Warn, "screenshot", firstLine(err.Error()))
	} else {
		visual(&res, shot, opt, add)
	}
	return res, pa.Text
}

func keyboardAudit(ctx context.Context, add func(sev, rule, detail string), clickNoKbd []string, posTab int) int {
	var focusable int
	_ = chromedp.Run(ctx, chromedp.Evaluate(countFocusable, &focusable))
	for _, c := range clickNoKbd {
		add(Warn, "click-not-keyboard", c+" reacts to clicks but cannot receive keyboard focus")
	}
	if posTab > 0 {
		add(Note, "positive-tabindex", fmt.Sprintf("%d element(s) use tabindex > 0, which overrides natural tab order", posTab))
	}
	if focusable == 0 {
		return 0
	}

	type probe struct {
		None      bool
		Key, Name string
		Indicator bool
		Visible   bool
	}
	seen := map[string]bool{}
	noIndicator := []string{}
	stops, sameRun, lastKey, none := 0, 0, "", 0
	for i := 0; i < 40; i++ {

		if err := chromedp.Run(ctx, chromedp.KeyEvent(kb.Tab), chromedp.Sleep(60*time.Millisecond)); err != nil {
			break
		}
		var raw string
		if err := chromedp.Run(ctx, chromedp.Evaluate(focusProbe, &raw)); err != nil {
			break
		}
		var p probe
		if json.Unmarshal([]byte(raw), &p) != nil {
			break
		}
		if p.None {
			none++
			continue
		}
		if p.Key == lastKey {
			sameRun++
			if sameRun >= 3 {
				add(Error, "keyboard-trap", "focus is stuck on "+p.Name+"; Tab cannot leave it")
				break
			}
			continue
		}
		sameRun, lastKey = 0, p.Key
		if !seen[p.Key] {
			seen[p.Key] = true
			stops++
			if !p.Indicator && len(noIndicator) < 8 {
				noIndicator = append(noIndicator, p.Name)
			}
		}
	}
	if stops == 0 {
		add(Error, "no-keyboard-focus", fmt.Sprintf("the page has %d focusable element(s) but Tab never focused any", focusable))
	}
	for _, n := range noIndicator {
		add(Warn, "no-focus-indicator", n+" shows no visible focus indicator (outline/box-shadow)")
	}
	return stops
}

var needsName = map[string]bool{"button": true, "link": true, "textbox": true, "combobox": true, "checkbox": true, "radio": true,
	"switch": true, "tab": true, "menuitem": true, "searchbox": true, "slider": true, "spinbutton": true}

func axAudit(ctx context.Context, add func(sev, rule, detail string)) {
	var nodes []*accessibility.Node
	err := chromedp.Run(ctx, chromedp.ActionFunc(func(c context.Context) error {
		if err := accessibility.Enable().Do(c); err != nil {
			return err
		}
		var err error
		nodes, err = accessibility.GetFullAXTree().Do(c)
		return err
	}))
	if err != nil {
		add(Note, "accessibility-tree", "unavailable: "+firstLine(err.Error()))
		return
	}
	var hasMain, hasH1 bool
	unnamed := map[string]int{}
	prev := 0
	for _, n := range nodes {
		if n.Ignored {
			continue
		}
		role, name := axString(n.Role), strings.TrimSpace(axString(n.Name))
		switch {
		case role == "main":
			hasMain = true
		case role == "heading":
			lvl := 0
			for _, p := range n.Properties {
				if p.Name == accessibility.PropertyNameLevel {
					fmt.Sscanf(string(p.Value.Value), "%d", &lvl)
				}
			}
			if lvl == 1 {
				hasH1 = true
			}
			if prev != 0 && lvl > prev+1 {
				add(Note, "heading-skip", fmt.Sprintf("heading level jumps from h%d to h%d", prev, lvl))
			}
			if lvl > 0 {
				prev = lvl
			}
		case needsName[role] && name == "":
			unnamed[role]++
		case role == "image" && name == "":
			unnamed["image"]++
		}
	}
	for role, n := range unnamed {
		sev := Error
		if role == "image" {
			sev = Warn
		}
		add(sev, "unnamed-"+role, fmt.Sprintf("%d %s element(s) have no accessible name; a screen reader announces them as just %q", n, role, role))
	}
	if !hasMain {
		add(Warn, "no-main-landmark", "no <main> landmark: screen-reader users cannot jump to the content")
	}
	if !hasH1 {
		add(Note, "no-h1-in-tree", "accessibility tree has no level-1 heading")
	}
}

var safeName = regexp.MustCompile(`[^a-zA-Z0-9]+`)

func fileKey(opt Options, pageURL, vp, loc string) string {
	return fmt.Sprintf("%s-%s-%s-%s", safeName.ReplaceAllString(opt.Name, "_"), safeName.ReplaceAllString(strings.TrimPrefix(strings.TrimPrefix(pageURL, "https://"), "http://"), "_"), vp, safeName.ReplaceAllString(loc, "_"))
}

func visual(res *Result, shot []byte, opt Options, add func(sev, rule, detail string)) {
	if opt.ArtifactDir == "" {
		return
	}
	key := fileKey(opt, res.Page, res.Viewport, res.Locale)
	cur := key + ".png"
	if err := os.WriteFile(filepath.Join(opt.ArtifactDir, cur), shot, 0o644); err == nil {
		res.Screenshot = cur
	}
	if opt.BaselineDir == "" {
		return
	}
	basePath := filepath.Join(opt.BaselineDir, key+".png")
	base, err := os.ReadFile(basePath)
	if err != nil || opt.UpdateBaseline {
		if werr := os.WriteFile(basePath, shot, 0o644); werr == nil {
			res.Baseline = "created"
			if opt.UpdateBaseline && err == nil {
				res.Baseline = "updated"
			}
		}
		return
	}
	res.Baseline = "compared"
	pct, same, diffPNG, derr := diffImages(base, shot)
	if derr != nil {
		add(Warn, "visual-regression", "cannot compare with baseline: "+derr.Error())
		return
	}
	res.DiffPct = &pct
	if diffPNG != nil {
		name := key + "-diff.png"
		if os.WriteFile(filepath.Join(opt.ArtifactDir, name), diffPNG, 0o644) == nil {
			res.DiffImage = name
		}
	}
	switch {
	case !same:
		add(Error, "visual-regression", "page size differs from the baseline (layout changed)")
	case pct > opt.DiffThresholdPct:
		add(Error, "visual-regression", fmt.Sprintf("%.2f%% of pixels differ from the baseline (limit %.2f%%)", pct, opt.DiffThresholdPct))
	}
}

func hostAllowed(raw string, allowed map[string]bool) bool {
	u, err := url.Parse(raw)
	if err != nil {
		return false
	}
	switch u.Scheme {
	case "data", "blob", "about":
		return true
	}
	return allowed[strings.ToLower(u.Hostname())]
}

func remoteText(o *runtime.RemoteObject) string {
	if o == nil {
		return ""
	}
	if len(o.Value) > 0 {
		var s string
		if json.Unmarshal([]byte(o.Value), &s) == nil {
			return s
		}
		return string(o.Value)
	}
	return o.Description
}

func axString(v *accessibility.Value) string {
	if v == nil {
		return ""
	}
	var s string
	if json.Unmarshal([]byte(v.Value), &s) == nil {
		return s
	}
	return strings.Trim(string(v.Value), `"`)
}

func firstLine(s string) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		s = s[:i]
	}
	if len(s) > 200 {
		s = s[:200] + "…"
	}
	return s
}

func orOne(f float64) float64 {
	if f <= 0 {
		return 1
	}
	return f
}
