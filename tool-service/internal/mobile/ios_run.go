package mobile

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

var (
	bundleRe   = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9.\-]*\.[A-Za-z0-9.\-]+$`)
	deviceRe   = regexp.MustCompile(`^[A-Za-z0-9 .()\-]{1,60}$`)
	execNameRe = regexp.MustCompile(`^[A-Za-z0-9_. \-]{1,80}$`)
	deepLinkRe = regexp.MustCompile("^[A-Za-z][A-Za-z0-9+.\\-]*://[^\\s;&|$`\"'\\\\]*$")
	pidRe      = regexp.MustCompile(`:\s*(\d+)\s*$`)
)

func ValidBundleID(s string) bool { return bundleRe.MatchString(s) }

func FindSimctl() error {
	if _, err := exec.LookPath("xcrun"); err != nil {
		return errors.New("iOS simulator testing needs macOS with Xcode (xcrun not found); this host cannot run it")
	}
	if err := exec.Command("xcrun", "-f", "simctl").Run(); err != nil {
		return errors.New("xcrun simctl is not available: install full Xcode (not only the command line tools) and run `xcode-select -s /Applications/Xcode.app`")
	}
	return nil
}

type ExecSimctl struct{}

func (ExecSimctl) Run(ctx context.Context, args ...string) ([]byte, error) {
	timeout := 90 * time.Second
	if len(args) > 0 && (args[0] == "bootstatus" || args[0] == "install") {
		timeout = 5 * time.Minute
	}
	cctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	var out bytes.Buffer
	cmd := exec.CommandContext(cctx, "xcrun", append([]string{"simctl"}, args...)...)
	cmd.Stdout, cmd.Stderr = &out, &out
	err := cmd.Run()
	return out.Bytes(), err
}

type Simulator struct {
	UDID    string `json:"udid"`
	Name    string `json:"name"`
	State   string `json:"state"`
	Runtime string `json:"runtime"`
}

func ParseSimulators(js []byte) ([]Simulator, error) {
	var v struct {
		Devices map[string][]struct {
			UDID, Name, State string
			IsAvailable       bool `json:"isAvailable"`
		} `json:"devices"`
	}
	if err := json.Unmarshal(js, &v); err != nil {
		return nil, fmt.Errorf("unexpected simctl output: %w", err)
	}
	var out []Simulator
	for rt, list := range v.Devices {
		if !strings.Contains(rt, "iOS") {
			continue
		}
		for _, d := range list {
			if d.IsAvailable {
				out = append(out, Simulator{d.UDID, d.Name, d.State, strings.TrimPrefix(rt, "com.apple.CoreSimulator.SimRuntime.")})
			}
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if (out[i].State == "Booted") != (out[j].State == "Booted") {
			return out[i].State == "Booted"
		}
		if out[i].Runtime != out[j].Runtime {
			return out[i].Runtime > out[j].Runtime
		}
		return out[i].Name < out[j].Name
	})
	return out, nil
}

func ListSimulators(ctx context.Context, r Runner) ([]Simulator, error) {
	out, err := r.Run(ctx, "list", "devices", "available", "-j")
	if err != nil {
		return nil, fmt.Errorf("simctl list failed: %w: %s", err, strings.TrimSpace(string(out)))
	}
	return ParseSimulators(out)
}

type IOSOptions struct {
	BundleID    string
	AppPath     string
	Device      string
	ExecName    string
	Scenarios   []string
	DeepLinks   []string
	ArtifactDir string
	CrashDir    string
	Sleep       func(time.Duration)
	Progress    func(string)
}

type IOSCrash struct {
	File    string `json:"file"`
	Summary string `json:"summary"`
}

type IOSReport struct {
	Device     Simulator        `json:"device"`
	Bundle     string           `json:"bundle_id"`
	Installed  bool             `json:"installed_by_test"`
	PID        string           `json:"pid,omitempty"`
	Screenshot string           `json:"screenshot,omitempty"`
	Scenarios  []ScenarioResult `json:"scenarios,omitempty"`
	Crashes    []IOSCrash       `json:"crashes,omitempty"`
	LogFaults  int              `json:"log_faults"`
	Findings   []Finding        `json:"findings"`
}

var AllIOSScenarios = []string{"dark_mode", "dynamic_type", "rtl_locale", "relaunch"}

type runIOS struct {
	r     Runner
	opt   IOSOptions
	rep   *IOSReport
	c     collector
	start time.Time
	seenC map[string]bool
}

func (a *runIOS) sleep(d time.Duration) {
	if a.opt.Sleep != nil {
		a.opt.Sleep(d)
		return
	}
	time.Sleep(d)
}

func (a *runIOS) run(ctx context.Context, args ...string) (string, error) {
	out, err := a.r.Run(ctx, args...)
	return strings.TrimSpace(string(out)), err
}

func (a *runIOS) launch(ctx context.Context, extra ...string) (string, error) {
	_, _ = a.run(ctx, "terminate", a.rep.Device.UDID, a.opt.BundleID)
	args := append([]string{"launch", a.rep.Device.UDID, a.opt.BundleID}, extra...)
	out, err := a.run(ctx, args...)
	if err != nil {
		return "", fmt.Errorf("%s", firstMatchLine(out, "error", "Error", "failed"))
	}
	if m := pidRe.FindStringSubmatch(out); m != nil {
		return m[1], nil
	}
	return "", nil
}

func (a *runIOS) alive(ctx context.Context) bool {
	out, _ := a.run(ctx, "spawn", a.rep.Device.UDID, "launchctl", "list")
	for _, l := range strings.Split(out, "\n") {
		if strings.Contains(l, "UIKitApplication:"+a.opt.BundleID) {
			f := strings.Fields(l)
			return len(f) > 0 && f[0] != "-"
		}
	}
	return false
}

func (a *runIOS) screenshot(ctx context.Context, name string) string {
	if a.opt.ArtifactDir == "" {
		return ""
	}
	file := fmt.Sprintf("ios-%s-%s-%d.png", strings.ReplaceAll(a.opt.BundleID, ".", "_"), name, time.Now().Unix())
	if _, err := a.run(ctx, "io", a.rep.Device.UDID, "screenshot", filepath.Join(a.opt.ArtifactDir, file)); err != nil {
		return ""
	}
	if _, err := os.Stat(filepath.Join(a.opt.ArtifactDir, file)); err != nil {
		return ""
	}
	return file
}

var (
	excTypeRe = regexp.MustCompile(`"exception"\s*:\s*\{[^}]*?"type"\s*:\s*"([A-Z_]+)"`)
	sigRe     = regexp.MustCompile(`"signal"\s*:\s*"([A-Z_ ()a-z]+)"`)
	termRe    = regexp.MustCompile(`"termination"\s*:\s*\{[^}]*?"indicator"\s*:\s*"([^"]+)"`)
)

func (a *runIOS) crashes(where string) []IOSCrash {
	dir := a.opt.CrashDir
	if dir == "" {
		if home, err := os.UserHomeDir(); err == nil {
			dir = filepath.Join(home, "Library", "Logs", "DiagnosticReports")
		}
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var fresh []IOSCrash
	for _, e := range entries {
		n := e.Name()
		if e.IsDir() || !strings.HasPrefix(n, a.opt.ExecName) || !(strings.HasSuffix(n, ".ips") || strings.HasSuffix(n, ".crash")) || a.seenC[n] {
			continue
		}
		info, err := e.Info()
		if err != nil || info.ModTime().Before(a.start) {
			continue
		}
		a.seenC[n] = true
		b, _ := os.ReadFile(filepath.Join(dir, n))
		if len(b) > 2<<20 {
			b = b[:2<<20]
		}
		var parts []string
		if m := excTypeRe.FindSubmatch(b); m != nil {
			parts = append(parts, string(m[1]))
		}
		if m := sigRe.FindSubmatch(b); m != nil {
			parts = append(parts, string(m[1]))
		}
		if m := termRe.FindSubmatch(b); m != nil {
			parts = append(parts, string(m[1]))
		}
		summary := strings.Join(parts, ", ")
		if summary == "" {
			summary = "crash report written"
		}
		c := IOSCrash{File: n, Summary: summary + " (during " + where + ")"}
		fresh = append(fresh, c)
		a.rep.Crashes = append(a.rep.Crashes, c)
		a.c.add("MOB-IOS-RT-001", High, "app crashed: "+c.Summary, n)
	}
	return fresh
}

func RunIOS(ctx context.Context, r Runner, opt IOSOptions) (*IOSReport, error) {
	if !ValidBundleID(opt.BundleID) {
		return nil, errors.New("invalid bundle id")
	}
	if opt.Device != "" && !udidRe.MatchString(opt.Device) && !deviceRe.MatchString(opt.Device) {
		return nil, errors.New("invalid device (use a simulator name or UDID)")
	}
	if opt.ExecName == "" {
		opt.ExecName = opt.BundleID[strings.LastIndex(opt.BundleID, ".")+1:]
	}
	if !execNameRe.MatchString(opt.ExecName) {
		return nil, errors.New("invalid executable name")
	}
	for _, l := range opt.DeepLinks {
		if !deepLinkRe.MatchString(l) {
			return nil, fmt.Errorf("invalid deep link %q", l)
		}
	}
	if opt.ArtifactDir != "" {
		if err := os.MkdirAll(opt.ArtifactDir, 0o755); err != nil {
			return nil, err
		}
	}

	sims, err := ListSimulators(ctx, r)
	if err != nil {
		return nil, err
	}
	var dev *Simulator
	for i := range sims {
		s := &sims[i]
		if opt.Device == "" || s.UDID == opt.Device || s.Name == opt.Device {
			dev = s
			break
		}
	}
	if dev == nil {
		return nil, errors.New("no matching iOS simulator found (create one in Xcode)")
	}
	a := &runIOS{r: r, opt: opt, rep: &IOSReport{Device: *dev, Bundle: opt.BundleID}, start: time.Now().Add(-2 * time.Second), seenC: map[string]bool{}}
	udid := dev.UDID

	if dev.State != "Booted" {
		if a.opt.Progress != nil {
			a.opt.Progress("booting simulator")
		}
		if out, err := a.run(ctx, "boot", udid); err != nil && !strings.Contains(out, "current state: Booted") {
			return nil, fmt.Errorf("cannot boot simulator: %s", firstMatchLine(out, "error", "Error", "Unable"))
		}
		if out, err := a.run(ctx, "bootstatus", udid, "-b"); err != nil {
			return nil, fmt.Errorf("simulator did not finish booting: %s", firstMatchLine(out, "error", "Error"))
		}
		a.rep.Device.State = "Booted"
	}

	if opt.AppPath != "" {
		if a.opt.Progress != nil {
			a.opt.Progress("installing")
		}
		if out, err := a.run(ctx, "install", udid, opt.AppPath); err != nil {
			return nil, fmt.Errorf("install failed (the .app must be built for the simulator, not a device .ipa): %s", firstMatchLine(out, "error", "Error", "Unable"))
		}
		a.rep.Installed = true
	} else if out, _ := a.run(ctx, "get_app_container", udid, opt.BundleID); out == "" || strings.Contains(out, "No such") {
		return nil, fmt.Errorf("%s is not installed on the simulator; upload a simulator .app", opt.BundleID)
	}

	if a.opt.Progress != nil {
		a.opt.Progress("launching")
	}
	pid, err := a.launch(ctx)
	if err != nil {
		a.c.add("MOB-IOS-LAUNCH-001", High, "app failed to launch", err.Error())
		a.rep.Findings = a.c.sorted()
		return a.rep, nil
	}
	a.rep.PID = pid
	a.sleep(3 * time.Second)
	a.rep.Screenshot = a.screenshot(ctx, "launch")
	if !a.alive(ctx) {
		a.c.add("MOB-IOS-LAUNCH-002", High, "app exited right after launching", "The process was not running 3 seconds after launch.")
	}
	a.crashes("launch")

	scenarios := opt.Scenarios
	if len(scenarios) == 0 {
		scenarios = AllIOSScenarios
	}
	for _, s := range scenarios {
		if a.opt.Progress != nil {
			a.opt.Progress("scenario: " + s)
		}
		if res, ok := a.scenario(ctx, s); ok {
			a.rep.Scenarios = append(a.rep.Scenarios, res)
		}
	}
	for _, link := range opt.DeepLinks {
		if a.opt.Progress != nil {
			a.opt.Progress("deep link " + link)
		}
		res := ScenarioResult{Name: "deep_link " + link}
		if out, err := a.run(ctx, "openurl", udid, link); err != nil {
			res.Problems = append(res.Problems, "not handled: "+firstMatchLine(out, "error", "Error", "OSStatus"))
			a.c.add("MOB-IOS-DL-001", Medium, "deep link is not handled", link)
		}
		a.sleep(2 * time.Second)
		res.Screenshot = a.screenshot(ctx, "deeplink")
		if !a.alive(ctx) {
			res.Problems = append(res.Problems, "app is not running afterwards")
			a.c.add("MOB-IOS-SCN-001", High, "app is not running after deep link "+link, "")
		}
		for _, c := range a.crashes("deep link " + link) {
			res.Problems = append(res.Problems, c.Summary)
		}
		res.Passed = len(res.Problems) == 0
		a.rep.Scenarios = append(a.rep.Scenarios, res)
	}

	logs, _ := a.run(ctx, "spawn", udid, "log", "show", "--last", "5m", "--style", "compact", "--predicate",
		fmt.Sprintf(`process == "%s" AND messageType == fault`, opt.ExecName))
	for _, l := range strings.Split(logs, "\n") {
		if strings.TrimSpace(l) != "" && !strings.HasPrefix(l, "Timestamp") && !strings.HasPrefix(l, "Filtering") {
			a.rep.LogFaults++
		}
	}
	if a.rep.LogFaults > 0 {
		a.c.add("MOB-IOS-LOG-001", Low, fmt.Sprintf("%d fault-level log message(s) from the app", a.rep.LogFaults), "Inspect with: xcrun simctl spawn booted log show --predicate 'messageType == fault'")
	}
	a.rep.Findings = a.c.sorted()
	return a.rep, nil
}

func (a *runIOS) scenario(ctx context.Context, name string) (ScenarioResult, bool) {
	res := ScenarioResult{Name: name}
	udid := a.rep.Device.UDID
	var restore func()
	var args []string
	switch name {
	case "dark_mode":
		prev, _ := a.run(ctx, "ui", udid, "appearance")
		_, _ = a.run(ctx, "ui", udid, "appearance", "dark")
		restore = func() { _, _ = a.run(ctx, "ui", udid, "appearance", oneOf(prev, "light", "dark")) }
	case "dynamic_type":
		prev, _ := a.run(ctx, "ui", udid, "content_size")
		_, _ = a.run(ctx, "ui", udid, "content_size", "accessibility-extra-extra-extra-large")
		restore = func() { _, _ = a.run(ctx, "ui", udid, "content_size", oneOfContent(prev)) }
	case "rtl_locale":
		args = []string{"-AppleLanguages", "(ar)", "-AppleLocale", "ar_SA"}
	case "relaunch":
	default:
		return res, false
	}
	if restore != nil {
		defer restore()
	}
	if _, err := a.launch(ctx, args...); err != nil {
		res.Problems = append(res.Problems, "launch failed: "+err.Error())
	}
	a.sleep(3 * time.Second)
	res.Screenshot = a.screenshot(ctx, name)
	if !a.alive(ctx) {
		res.Problems = append(res.Problems, "app is not running afterwards")
		a.c.add("MOB-IOS-SCN-001", High, "app is not running after scenario: "+name, "")
	}
	for _, c := range a.crashes("scenario " + name) {
		res.Problems = append(res.Problems, c.Summary)
	}
	res.Passed = len(res.Problems) == 0
	return res, true
}

func oneOf(v string, allowed ...string) string {
	for _, a := range allowed {
		if v == a {
			return v
		}
	}
	return allowed[0]
}

func oneOfContent(v string) string {
	valid := map[string]bool{"extra-small": true, "small": true, "medium": true, "large": true, "extra-large": true, "extra-extra-large": true, "extra-extra-extra-large": true,
		"accessibility-medium": true, "accessibility-large": true, "accessibility-extra-large": true, "accessibility-extra-extra-large": true, "accessibility-extra-extra-extra-large": true}
	if valid[v] {
		return v
	}
	return "medium"
}
