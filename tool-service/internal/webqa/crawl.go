package webqa

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/net/html"
)

const (
	Error = "error"
	Warn  = "warn"
	Note  = "info"
)

type Issue struct {
	Severity string `json:"severity"`
	Rule     string `json:"rule"`
	Detail   string `json:"detail,omitempty"`
}

type Page struct {
	URL       string  `json:"url"`
	Status    int     `json:"status"`
	LatencyMs int64   `json:"latency_ms"`
	SizeKB    float64 `json:"size_kb"`
	Title     string  `json:"title,omitempty"`
	Issues    []Issue `json:"issues,omitempty"`
}

type BrokenLink struct {
	From   string `json:"from"`
	URL    string `json:"url"`
	Status int    `json:"status"`
	Error  string `json:"error,omitempty"`
}

type Field struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	Value string `json:"value,omitempty"`
}

type Form struct {
	Page   string  `json:"page"`
	Method string  `json:"method"`
	Action string  `json:"action"`
	Fields []Field `json:"fields"`
}

type Report struct {
	Forms       []Form       `json:"forms,omitempty"`
	Pages       []Page       `json:"pages"`
	BrokenLinks []BrokenLink `json:"broken_links"`
	Errors      int          `json:"error_count"`
	Warnings    int          `json:"warning_count"`
}

const maxBody = 2 << 20

func Crawl(ctx context.Context, c *http.Client, start string, maxPages, depth int) Report {
	if maxPages <= 0 || maxPages > 100 {
		maxPages = 20
	}
	su, err := url.Parse(start)
	rep := Report{}
	if err != nil || su.Host == "" {
		return rep
	}
	type node struct {
		u string
		d int
	}
	queue := []node{{normalize(su, su), 0}}
	seen := map[string]bool{queue[0].u: true}
	referrer := map[string]string{}

	for len(queue) > 0 && len(rep.Pages) < maxPages && ctx.Err() == nil {
		n := queue[0]
		queue = queue[1:]
		pu, _ := url.Parse(n.u)

		rctx, cancel := context.WithTimeout(ctx, 15*time.Second)
		req, _ := http.NewRequestWithContext(rctx, http.MethodGet, n.u, nil)
		t0 := time.Now()
		resp, err := c.Do(req)
		if err != nil {
			cancel()
			if from := referrer[n.u]; from != "" {
				rep.BrokenLinks = append(rep.BrokenLinks, BrokenLink{From: from, URL: n.u, Error: shorten(err.Error())})
			} else {
				rep.Pages = append(rep.Pages, Page{URL: n.u, Issues: []Issue{{Error, "unreachable", shorten(err.Error())}}})
				rep.Errors++
			}
			continue
		}
		body, _ := io.ReadAll(io.LimitReader(resp.Body, maxBody))
		resp.Body.Close()
		cancel()
		lat := time.Since(t0)

		pg := Page{URL: n.u, Status: resp.StatusCode, LatencyMs: lat.Milliseconds(), SizeKB: float64(len(body)) / 1024}
		if resp.StatusCode >= 400 {
			if from := referrer[n.u]; from != "" {
				rep.BrokenLinks = append(rep.BrokenLinks, BrokenLink{From: from, URL: n.u, Status: resp.StatusCode})
				continue
			}
			pg.Issues = append(pg.Issues, Issue{Error, "http-status", fmt.Sprintf("start page returned %d", resp.StatusCode)})
		}
		if lat > 1500*time.Millisecond {
			pg.Issues = append(pg.Issues, Issue{Warn, "slow-response", fmt.Sprintf("%dms", lat.Milliseconds())})
		}
		if len(body) >= maxBody {
			pg.Issues = append(pg.Issues, Issue{Warn, "page-weight", "HTML larger than 2 MB"})
		}

		if strings.Contains(resp.Header.Get("Content-Type"), "html") {
			doc, err := html.Parse(strings.NewReader(string(body)))
			if err == nil {
				title, links, issues := analyze(doc, pu)
				pg.Title = title
				if len(rep.Forms) < 30 {
					rep.Forms = append(rep.Forms, extractForms(doc, pu)...)
				}
				pg.Issues = append(pg.Issues, issues...)
				if n.d < depth {
					for _, l := range links {
						if !seen[l] {
							seen[l] = true
							referrer[l] = n.u
							queue = append(queue, node{l, n.d + 1})
						}
					}
				}
			}
		}
		for _, is := range pg.Issues {
			if is.Severity == Error {
				rep.Errors++
			} else if is.Severity == Warn {
				rep.Warnings++
			}
		}
		rep.Pages = append(rep.Pages, pg)
	}
	rep.Errors += len(rep.BrokenLinks)
	return rep
}

func analyze(doc *html.Node, page *url.URL) (string, []string, []Issue) {
	var (
		title, lang        string
		hasViewport, hasH1 bool
		h1s                int
		lastHeading        int
		links              []string
		issues             []Issue
		ids                = map[string]int{}
		labelFor           = map[string]bool{}
		fields             []*html.Node
	)
	add := func(sev, rule, detail string) { issues = append(issues, Issue{sev, rule, detail}) }
	isHTTPS := page.Scheme == "https"

	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode {
			if id := attr(n, "id"); id != "" {
				ids[id]++
			}
			switch n.Data {
			case "html":
				lang = attr(n, "lang")
			case "title":
				title = strings.TrimSpace(text(n))
			case "meta":
				if strings.EqualFold(attr(n, "name"), "viewport") {
					hasViewport = true
				}
			case "img":
				if _, ok := hasAttr(n, "alt"); !ok && attr(n, "role") != "presentation" {
					add(Warn, "img-alt", "image without alt: "+shorten(attr(n, "src")))
				}
				if isHTTPS && strings.HasPrefix(attr(n, "src"), "http://") {
					add(Warn, "mixed-content", "insecure image "+shorten(attr(n, "src")))
				}
			case "script":
				if isHTTPS && strings.HasPrefix(attr(n, "src"), "http://") {
					add(Error, "mixed-content", "insecure script "+shorten(attr(n, "src")))
				}
			case "label":
				if f := attr(n, "for"); f != "" {
					labelFor[f] = true
				}
			case "input", "select", "textarea":
				t := strings.ToLower(attr(n, "type"))
				if n.Data == "input" && (t == "hidden" || t == "submit" || t == "button" || t == "reset" || t == "image") {
					break
				}
				fields = append(fields, n)
			case "button":
				if strings.TrimSpace(text(n)) == "" && attr(n, "aria-label") == "" && attr(n, "aria-labelledby") == "" && attr(n, "title") == "" {
					add(Warn, "button-name", "button has no accessible name")
				}
			case "a":
				href := strings.TrimSpace(attr(n, "href"))
				if strings.TrimSpace(text(n)) == "" && attr(n, "aria-label") == "" && !hasImgAlt(n) {
					add(Warn, "link-name", "link has no accessible text: "+shorten(href))
				}
				if attr(n, "target") == "_blank" && !strings.Contains(attr(n, "rel"), "noopener") {
					add(Note, "noopener", "target=_blank without rel=noopener: "+shorten(href))
				}
				if l, ok := sameHostLink(page, href); ok {
					links = append(links, l)
				}
			case "h1", "h2", "h3", "h4", "h5", "h6":
				lvl := int(n.Data[1] - '0')
				if lvl == 1 {
					hasH1 = true
					h1s++
				}
				if lastHeading != 0 && lvl > lastHeading+1 {
					add(Note, "heading-order", fmt.Sprintf("h%d follows h%d", lvl, lastHeading))
				}
				lastHeading = lvl
			}
		}
		for ch := n.FirstChild; ch != nil; ch = ch.NextSibling {
			walk(ch)
		}
	}
	walk(doc)

	if lang == "" {
		add(Warn, "html-lang", "<html> has no lang attribute")
	}
	if title == "" {
		add(Warn, "title", "page has no <title>")
	}
	if !hasViewport {
		add(Warn, "viewport", "no meta viewport (not mobile friendly)")
	}
	if !hasH1 {
		add(Note, "h1", "page has no <h1>")
	} else if h1s > 1 {
		add(Note, "h1", fmt.Sprintf("page has %d <h1> elements", h1s))
	}
	for id, n := range ids {
		if n > 1 {
			add(Error, "duplicate-id", fmt.Sprintf("id %q used %d times", id, n))
		}
	}
	for _, f := range fields {
		id := attr(f, "id")
		if attr(f, "aria-label") == "" && attr(f, "aria-labelledby") == "" && attr(f, "title") == "" && !(id != "" && labelFor[id]) && !insideLabel(f) {
			add(Warn, "form-label", fmt.Sprintf("<%s name=%q> has no label", f.Data, attr(f, "name")))
		}
	}
	return title, links, issues
}

func sameHostLink(page *url.URL, href string) (string, bool) {
	if href == "" || strings.HasPrefix(href, "#") {
		return "", false
	}
	l := strings.ToLower(href)
	for _, p := range []string{"mailto:", "tel:", "javascript:", "data:", "sms:"} {
		if strings.HasPrefix(l, p) {
			return "", false
		}
	}
	u, err := page.Parse(href)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host != page.Host {
		return "", false
	}
	return normalize(u, page), true
}

func normalize(u, base *url.URL) string {
	c := *u
	c.Fragment = ""
	if c.Path == "" {
		c.Path = "/"
	}
	return c.String()
}

func attr(n *html.Node, k string) string {
	v, _ := hasAttr(n, k)
	return v
}

func hasAttr(n *html.Node, k string) (string, bool) {
	for _, a := range n.Attr {
		if a.Key == k {
			return a.Val, true
		}
	}
	return "", false
}

func text(n *html.Node) string {
	var b strings.Builder
	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.TextNode {
			b.WriteString(n.Data)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(n)
	return b.String()
}

func hasImgAlt(n *html.Node) bool {
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode && c.Data == "img" && strings.TrimSpace(attr(c, "alt")) != "" {
			return true
		}
		if hasImgAlt(c) {
			return true
		}
	}
	return false
}

func insideLabel(n *html.Node) bool {
	for p := n.Parent; p != nil; p = p.Parent {
		if p.Type == html.ElementNode && p.Data == "label" {
			return true
		}
	}
	return false
}

func shorten(s string) string {
	if len(s) > 100 {
		return s[:100] + "…"
	}
	return s
}

func extractForms(doc *html.Node, page *url.URL) []Form {
	var out []Form
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "form" {
			action := page
			if a := attr(n, "action"); a != "" {
				if u, err := page.Parse(a); err == nil {
					action = u
				}
			}
			if action.Host == page.Host && (action.Scheme == "http" || action.Scheme == "https") {
				f := Form{Page: page.String(), Method: strings.ToUpper(orDefault(attr(n, "method"), "GET")), Action: normalize(action, page)}
				var fields func(*html.Node)
				fields = func(c *html.Node) {
					if c.Type == html.ElementNode && (c.Data == "input" || c.Data == "textarea" || c.Data == "select") && attr(c, "name") != "" {
						t := strings.ToLower(orDefault(attr(c, "type"), "text"))
						if c.Data != "input" {
							t = c.Data
						}
						if t != "submit" && t != "button" && t != "image" && t != "reset" && t != "file" {
							f.Fields = append(f.Fields, Field{Name: attr(c, "name"), Type: t, Value: attr(c, "value")})
						}
					}
					for ch := c.FirstChild; ch != nil; ch = ch.NextSibling {
						fields(ch)
					}
				}
				fields(n)
				if len(f.Fields) > 0 {
					out = append(out, f)
				}
			}
		}
		for ch := n.FirstChild; ch != nil; ch = ch.NextSibling {
			walk(ch)
		}
	}
	walk(doc)
	return out
}

func orDefault(s, d string) string {
	if s == "" {
		return d
	}
	return s
}
