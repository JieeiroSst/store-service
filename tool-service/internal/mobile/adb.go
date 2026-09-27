package mobile

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Runner interface {
	Run(ctx context.Context, args ...string) ([]byte, error)
}

var (
	pkgRe    = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_]*(\.[A-Za-z][A-Za-z0-9_]*)+$`)
	compRe   = regexp.MustCompile(`^[A-Za-z0-9_.$]+/[A-Za-z0-9_.$]+$`)
	serialRe = regexp.MustCompile(`^[A-Za-z0-9._:\-]{1,80}$`)
	udidRe   = regexp.MustCompile(`^[A-Fa-f0-9\-]{36}$`)
)

func ValidPackage(s string) bool   { return pkgRe.MatchString(s) }
func ValidComponent(s string) bool { return compRe.MatchString(s) }
func ValidSerial(s string) bool    { return serialRe.MatchString(s) }

func FindADB() (string, error) {
	if p := os.Getenv("ADB_PATH"); p != "" {
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}
	if p, err := exec.LookPath("adb"); err == nil {
		return p, nil
	}
	return "", errors.New("adb not found: install Android platform-tools or set ADB_PATH")
}

type ExecADB struct {
	Bin, Serial string
}

func (r ExecADB) Run(ctx context.Context, args ...string) ([]byte, error) {
	full := args
	if r.Serial != "" {
		full = append([]string{"-s", r.Serial}, args...)
	}
	timeout := 90 * time.Second
	if len(args) > 0 && (args[0] == "install" || args[0] == "wait-for-device") {
		timeout = 5 * time.Minute
	}
	if len(args) > 1 && args[0] == "shell" && args[1] == "monkey" {
		timeout = 10 * time.Minute
	}
	cctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	var out bytes.Buffer
	cmd := exec.CommandContext(cctx, r.Bin, full...)
	cmd.Stdout, cmd.Stderr = &out, &out
	err := cmd.Run()
	return out.Bytes(), err
}

type Device struct {
	Serial  string `json:"serial"`
	State   string `json:"state"`
	Model   string `json:"model,omitempty"`
	Product string `json:"product,omitempty"`
	Release string `json:"android_version,omitempty"`
	SDK     int    `json:"sdk,omitempty"`
	Screen  string `json:"screen,omitempty"`
	Density int    `json:"density_dpi,omitempty"`
}

func ParseDevices(out string) []Device {
	ds := []Device{}
	for _, line := range strings.Split(out, "\n") {
		f := strings.Fields(line)
		if len(f) < 2 || f[0] == "List" || strings.HasPrefix(f[0], "*") {
			continue
		}
		d := Device{Serial: f[0], State: f[1]}
		for _, kv := range f[2:] {
			if k, v, ok := strings.Cut(kv, ":"); ok {
				switch k {
				case "model":
					d.Model = v
				case "product":
					d.Product = v
				}
			}
		}
		ds = append(ds, d)
	}
	return ds
}

func ListAndroidDevices(ctx context.Context, r Runner) ([]Device, error) {
	out, err := r.Run(ctx, "devices", "-l")
	if err != nil {
		return nil, fmt.Errorf("adb devices failed: %w: %s", err, strings.TrimSpace(string(out)))
	}
	return ParseDevices(string(out)), nil
}

type Startup struct {
	ColdMs   []int `json:"cold_ms"`
	MedianMs int   `json:"median_ms"`
}

var (
	totalTimeRe = regexp.MustCompile(`TotalTime:\s*(\d+)`)
	pssNewRe    = regexp.MustCompile(`TOTAL PSS:\s*(\d+)`)
	pssOldRe    = regexp.MustCompile(`(?m)^\s*TOTAL\s+(\d+)`)
	jankRe      = regexp.MustCompile(`Janky frames:\s*(\d+)\s*\(([\d.]+)%\)`)
	framesRe    = regexp.MustCompile(`Total frames rendered:\s*(\d+)`)
	pctRe       = regexp.MustCompile(`(\d+)th percentile:\s*(\d+)ms`)
	densityRe   = regexp.MustCompile(`(?:Override|Physical) density:\s*(\d+)`)
	boundsRe    = regexp.MustCompile(`\[(\d+),(\d+)\]\[(\d+),(\d+)\]`)
)

func ParseStartTime(out string) (int, error) {
	if strings.Contains(out, "Error:") || strings.Contains(out, "Status: timeout") {
		return 0, fmt.Errorf("launch failed: %s", firstMatchLine(out, "Error:", "Status:"))
	}
	m := totalTimeRe.FindStringSubmatch(out)
	if m == nil {
		return 0, errors.New("no TotalTime in am start output")
	}
	n, _ := strconv.Atoi(m[1])
	return n, nil
}

func firstMatchLine(s string, subs ...string) string {
	for _, l := range strings.Split(s, "\n") {
		for _, sub := range subs {
			if strings.Contains(l, sub) {
				return strings.TrimSpace(l)
			}
		}
	}
	return strings.TrimSpace(s)
}

func median(v []int) int {
	if len(v) == 0 {
		return 0
	}
	c := append([]int(nil), v...)
	sort.Ints(c)
	return c[len(c)/2]
}

type MemInfo struct {
	PSSMB float64 `json:"pss_mb"`
}

func ParseMeminfo(out string) (float64, bool) {
	if m := pssNewRe.FindStringSubmatch(out); m != nil {
		kb, _ := strconv.Atoi(m[1])
		return float64(kb) / 1024, true
	}
	if m := pssOldRe.FindStringSubmatch(out); m != nil {
		kb, _ := strconv.Atoi(m[1])
		return float64(kb) / 1024, true
	}
	return 0, false
}

type Frames struct {
	Total    int     `json:"total_frames"`
	Janky    int     `json:"janky_frames"`
	JankyPct float64 `json:"janky_pct"`
	P90Ms    int     `json:"p90_ms,omitempty"`
	P99Ms    int     `json:"p99_ms,omitempty"`
}

func ParseGfxinfo(out string) (Frames, bool) {
	var f Frames
	if m := framesRe.FindStringSubmatch(out); m != nil {
		f.Total, _ = strconv.Atoi(m[1])
	} else {
		return f, false
	}
	if m := jankRe.FindStringSubmatch(out); m != nil {
		f.Janky, _ = strconv.Atoi(m[1])
		f.JankyPct, _ = strconv.ParseFloat(m[2], 64)
	}
	for _, m := range pctRe.FindAllStringSubmatch(out, -1) {
		v, _ := strconv.Atoi(m[2])
		switch m[1] {
		case "90":
			f.P90Ms = v
		case "99":
			f.P99Ms = v
		}
	}
	return f, true
}

type Crash struct {
	Kind    string   `json:"kind"`
	Summary string   `json:"summary"`
	Stack   []string `json:"stack,omitempty"`
}

func belongs(proc, pkg string) bool { return proc == pkg || strings.HasPrefix(proc, pkg+":") }

var (
	fatalRe    = regexp.MustCompile(`FATAL EXCEPTION`)
	processRe  = regexp.MustCompile(`Process:\s*([\w.:]+),\s*PID:\s*(\d+)`)
	anrRe      = regexp.MustCompile(`ANR in ([\w.:]+)`)
	nativeRe   = regexp.MustCompile(`Fatal signal (\d+) \((\w+)\).*pid \d+ \(([\w.:]+)\)`)
	strictRe   = regexp.MustCompile(`StrictMode policy violation`)
	logTagLine = regexp.MustCompile(`^\S+\s+\S+\s+\d+\s+\d+\s+([A-Z])\s+([^:]+?)\s*:\s?(.*)$`)
)

func stripLog(l string) (level, tag, msg string, ok bool) {
	m := logTagLine.FindStringSubmatch(l)
	if m == nil {
		return "", "", "", false
	}
	return m[1], strings.TrimSpace(m[2]), m[3], true
}

func ParseCrashes(logcat, pkg string) (crashes []Crash, strict int) {
	lines := strings.Split(logcat, "\n")
	for i := 0; i < len(lines); i++ {
		level, tag, msg, ok := stripLog(lines[i])
		if !ok {
			continue
		}
		switch {
		case tag == "AndroidRuntime" && fatalRe.MatchString(msg):
			var block []string
			for j := i + 1; j < len(lines) && len(block) < 40; j++ {
				_, t2, m2, ok2 := stripLog(lines[j])
				if !ok2 || t2 != "AndroidRuntime" {
					break
				}
				block = append(block, m2)
			}
			proc := ""
			for _, b := range block {
				if m := processRe.FindStringSubmatch(b); m != nil {
					proc = m[1]
					break
				}
			}
			if proc == "" || !belongs(proc, pkg) {
				continue
			}
			c := Crash{Kind: "java"}
			for _, b := range block {
				if processRe.MatchString(b) {
					continue
				}
				if c.Summary == "" && (strings.Contains(b, "Exception") || strings.Contains(b, "Error")) {
					c.Summary = strings.TrimSpace(b)
					continue
				}
				if strings.HasPrefix(strings.TrimSpace(b), "at ") && len(c.Stack) < 8 {
					c.Stack = append(c.Stack, strings.TrimSpace(b))
				}
			}
			if c.Summary == "" {
				c.Summary = "app crashed (no exception message captured)"
			}
			crashes = append(crashes, c)
		case level == "E" && anrRe.MatchString(msg):
			if m := anrRe.FindStringSubmatch(msg); belongs(m[1], pkg) {
				c := Crash{Kind: "anr", Summary: strings.TrimSpace(msg)}
				for j := i + 1; j < len(lines) && j < i+4; j++ {
					if _, _, m2, ok2 := stripLog(lines[j]); ok2 && strings.Contains(m2, "Reason:") {
						c.Summary += " — " + strings.TrimSpace(m2)
						break
					}
				}
				crashes = append(crashes, c)
			}
		case level == "F" && nativeRe.MatchString(msg):
			if m := nativeRe.FindStringSubmatch(msg); belongs(m[3], pkg) {
				crashes = append(crashes, Crash{Kind: "native", Summary: "native crash: signal " + m[1] + " (" + m[2] + ")"})
			}
		case strictRe.MatchString(msg):
			strict++
		}
	}
	return crashes, strict
}

type MonkeyResult struct {
	Events  int    `json:"events_injected"`
	Crashed bool   `json:"crashed"`
	ANR     bool   `json:"anr"`
	Aborted bool   `json:"aborted"`
	Summary string `json:"summary,omitempty"`
}

var eventsRe = regexp.MustCompile(`Events injected:\s*(\d+)`)

func ParseMonkey(out string) MonkeyResult {
	var r MonkeyResult
	if m := eventsRe.FindStringSubmatch(out); m != nil {
		r.Events, _ = strconv.Atoi(m[1])
	}
	for _, l := range strings.Split(out, "\n") {
		switch {
		case strings.Contains(l, "// CRASH:"):
			r.Crashed = true
		case strings.Contains(l, "// NOT RESPONDING:"):
			r.ANR = true
		case strings.Contains(l, "** Monkey aborted"):
			r.Aborted = true
		case strings.Contains(l, "// Short Msg:") && r.Summary == "":
			r.Summary = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(l), "// Short Msg:"))
		}
	}
	return r
}

type UIIssue struct {
	Rule   string `json:"rule"`
	Detail string `json:"detail"`
}

func AnalyzeUI(xmlDump []byte, pkg string, densityDPI int) ([]UIIssue, int, error) {
	root, err := parseTree(xmlDump)
	if err != nil {
		return nil, 0, fmt.Errorf("cannot parse uiautomator dump: %w", err)
	}
	if densityDPI <= 0 {
		densityDPI = 420
	}
	var issues []UIIssue
	counts := map[string]int{}
	total := 0
	var hasLabel func(n *node) bool
	hasLabel = func(n *node) bool {
		if strings.TrimSpace(n.attrs["text"]) != "" || strings.TrimSpace(n.attrs["content-desc"]) != "" {
			return true
		}
		for _, k := range n.kids {
			if hasLabel(k) {
				return true
			}
		}
		return false
	}
	var walk func(n *node)
	walk = func(n *node) {
		if n.name == "node" && n.attrs["package"] == pkg && n.attrs["clickable"] == "true" && n.attrs["enabled"] != "false" {
			total++
			id := n.attrs["resource-id"]
			if id == "" {
				id = n.attrs["class"]
			}
			if !hasLabel(n) || n.attrs["NAF"] == "true" {
				counts["unlabeled-control"]++
				if counts["unlabeled-control"] <= 10 {
					issues = append(issues, UIIssue{"unlabeled-control", id + " is clickable but has no text or content description (TalkBack cannot announce it)"})
				}
			}
			if m := boundsRe.FindStringSubmatch(n.attrs["bounds"]); m != nil {
				w, h := atoi(m[3])-atoi(m[1]), atoi(m[4])-atoi(m[2])
				wdp, hdp := w*160/densityDPI, h*160/densityDPI
				if w > 0 && h > 0 && (wdp < 48 || hdp < 48) {
					counts["small-touch-target"]++
					if counts["small-touch-target"] <= 10 {
						issues = append(issues, UIIssue{"small-touch-target", fmt.Sprintf("%s is %dx%d dp (minimum 48x48 dp)", id, wdp, hdp)})
					}
				}
			}
		}
		for _, k := range n.kids {
			walk(k)
		}
	}
	walk(root)
	return issues, total, nil
}

func atoi(s string) int { n, _ := strconv.Atoi(s); return n }

func ParseDensity(out string) int {
	if m := densityRe.FindStringSubmatch(out); m != nil {
		return atoi(m[1])
	}
	return 0
}
