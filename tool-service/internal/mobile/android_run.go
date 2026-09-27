package mobile

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type AndroidOptions struct {
	Package      string
	APKPath      string
	Component    string
	MonkeyEvents int
	Seed         int64
	Scenarios    []string
	ArtifactDir  string
	Uninstall    bool
	Sleep        func(time.Duration)
	Progress     func(string)
}

type ScenarioResult struct {
	Name       string   `json:"name"`
	Passed     bool     `json:"passed"`
	Screenshot string   `json:"screenshot,omitempty"`
	Problems   []string `json:"problems,omitempty"`
}

type AndroidReport struct {
	Device     Device           `json:"device"`
	Package    string           `json:"package"`
	Installed  bool             `json:"installed_by_test"`
	Startup    Startup          `json:"startup"`
	Memory     MemInfo          `json:"memory"`
	Frames     *Frames          `json:"frames,omitempty"`
	Monkey     *MonkeyResult    `json:"monkey,omitempty"`
	Crashes    []Crash          `json:"crashes,omitempty"`
	UI         []UIIssue        `json:"ui_issues,omitempty"`
	UIControls int              `json:"ui_clickable_controls"`
	Scenarios  []ScenarioResult `json:"scenarios,omitempty"`
	Screenshot string           `json:"screenshot,omitempty"`
	Findings   []Finding        `json:"findings"`
}

var AllAndroidScenarios = []string{"rotate", "font_scale", "dark_mode", "locale_rtl", "background"}

type runAndroid struct {
	r     Runner
	opt   AndroidOptions
	rep   *AndroidReport
	c     collector
	seen  map[string]bool
	crash map[string]bool
}

func (a *runAndroid) sh(ctx context.Context, args ...string) string {
	out, _ := a.r.Run(ctx, append([]string{"shell"}, args...)...)
	return strings.TrimSpace(string(out))
}

func (a *runAndroid) sleep(d time.Duration) {
	if a.opt.Sleep != nil {
		a.opt.Sleep(d)
		return
	}
	time.Sleep(d)
}

func (a *runAndroid) step(msg string) {
	if a.opt.Progress != nil {
		a.opt.Progress(msg)
	}
}

func (a *runAndroid) launch(ctx context.Context) (int, error) {
	pkg := a.opt.Package
	if a.opt.Component != "" {
		out, err := a.r.Run(ctx, "shell", "am", "start", "-W", "-n", a.opt.Component)
		if err != nil && len(out) == 0 {
			return 0, err
		}
		return ParseStartTime(string(out))
	}
	out, err := a.r.Run(ctx, "shell", "monkey", "-p", pkg, "-c", "android.intent.category.LAUNCHER", "1")
	if err != nil && len(out) == 0 {
		return 0, err
	}
	if strings.Contains(string(out), "No activities found") {
		return 0, errors.New("no launcher activity found for " + pkg)
	}
	return 0, nil
}

func (a *runAndroid) alive(ctx context.Context) bool {
	return a.sh(ctx, "pidof", a.opt.Package) != ""
}

func (a *runAndroid) collectCrashes(ctx context.Context, where string) []Crash {
	out, _ := a.r.Run(ctx, "logcat", "-d", "-v", "threadtime")
	cs, strict := ParseCrashes(string(out), a.opt.Package)
	_, _ = a.r.Run(ctx, "logcat", "-c")
	if strict > 0 && !a.seen["strict"] {
		a.seen["strict"] = true
		a.c.add("MOB-AND-STRICT-001", Low, fmt.Sprintf("%d StrictMode policy violation(s) logged", strict), "Disk or network work on the main thread, leaked resources.")
	}
	var fresh []Crash
	for _, c := range cs {
		key := c.Kind + c.Summary
		if a.crash[key] {
			continue
		}
		a.crash[key] = true
		c.Summary = strings.TrimSpace(c.Summary)
		if where != "" {
			c.Summary += " (during " + where + ")"
		}
		fresh = append(fresh, c)
		a.rep.Crashes = append(a.rep.Crashes, c)
		id, sev := "MOB-AND-RT-001", High
		switch c.Kind {
		case "anr":
			id = "MOB-AND-RT-002"
		case "native":
			id = "MOB-AND-RT-003"
		}
		a.c.add(id, sev, c.Kind+" crash: "+c.Summary, strings.Join(c.Stack, " | "))
	}
	return fresh
}

func (a *runAndroid) screenshot(ctx context.Context, name string) string {
	if a.opt.ArtifactDir == "" {
		return ""
	}
	out, err := a.r.Run(ctx, "exec-out", "screencap", "-p")
	if err != nil || len(out) < 8 || string(out[1:4]) != "PNG" {
		return ""
	}
	file := fmt.Sprintf("android-%s-%s-%d.png", strings.ReplaceAll(a.opt.Package, ".", "_"), name, time.Now().Unix())
	if os.WriteFile(filepath.Join(a.opt.ArtifactDir, file), out, 0o644) != nil {
		return ""
	}
	return file
}

func RunAndroid(ctx context.Context, r Runner, opt AndroidOptions) (*AndroidReport, error) {
	if !ValidPackage(opt.Package) {
		return nil, errors.New("invalid package name")
	}
	if opt.Component != "" && !ValidComponent(opt.Component) {
		return nil, errors.New("invalid component (expected package/Activity)")
	}
	if opt.MonkeyEvents < 0 || opt.MonkeyEvents > 20000 {
		return nil, errors.New("monkey_events must be between 0 and 20000")
	}
	if opt.ArtifactDir != "" {
		if err := os.MkdirAll(opt.ArtifactDir, 0o755); err != nil {
			return nil, err
		}
	}
	a := &runAndroid{r: r, opt: opt, rep: &AndroidReport{Package: opt.Package}, seen: map[string]bool{}, crash: map[string]bool{}}

	state := a.sh(ctx, "getprop", "sys.boot_completed")
	if state != "1" {
		return nil, errors.New("device is not ready (sys.boot_completed != 1)")
	}
	d := &a.rep.Device
	d.Model, d.Release = a.sh(ctx, "getprop", "ro.product.model"), a.sh(ctx, "getprop", "ro.build.version.release")
	d.SDK, _ = strconv.Atoi(a.sh(ctx, "getprop", "ro.build.version.sdk"))
	d.Screen = strings.TrimPrefix(a.sh(ctx, "wm", "size"), "Physical size: ")
	d.Density = ParseDensity(a.sh(ctx, "wm", "density"))

	if opt.APKPath != "" {
		a.step("installing")
		out, err := r.Run(ctx, "install", "-r", "-g", opt.APKPath)
		if err != nil || !strings.Contains(string(out), "Success") {
			return nil, fmt.Errorf("install failed: %s", firstMatchLine(string(out), "INSTALL_", "Failure", "Error"))
		}
		a.rep.Installed = true
		if opt.Uninstall {
			defer func() { _, _ = r.Run(context.Background(), "uninstall", opt.Package) }()
		}
	} else if !strings.Contains(a.sh(ctx, "pm", "list", "packages", opt.Package), "package:"+opt.Package) {
		return nil, fmt.Errorf("package %s is not installed on the device; upload an APK", opt.Package)
	}

	if opt.Component == "" {
		if line := lastLine(a.sh(ctx, "cmd", "package", "resolve-activity", "--brief", opt.Package)); ValidComponent(line) {
			a.opt.Component = line
		}
	}

	_, _ = r.Run(ctx, "logcat", "-c")
	a.step("cold start")
	for i := 0; i < 3; i++ {
		a.sh(ctx, "am", "force-stop", opt.Package)
		a.sleep(time.Second)
		ms, err := a.launch(ctx)
		if err != nil {
			a.c.add("MOB-AND-LAUNCH-001", High, "app failed to launch", err.Error())
			a.rep.Findings = a.c.sorted()
			return a.rep, nil
		}
		if ms > 0 {
			a.rep.Startup.ColdMs = append(a.rep.Startup.ColdMs, ms)
		}
		a.sleep(1500 * time.Millisecond)
	}
	a.rep.Startup.MedianMs = median(a.rep.Startup.ColdMs)
	switch m := a.rep.Startup.MedianMs; {
	case m > 5000:
		a.c.add("MOB-AND-PERF-001", Medium, fmt.Sprintf("slow cold start: %d ms", m), "Android vitals treats cold starts over 5 s as excessive.")
	case m > 3000:
		a.c.add("MOB-AND-PERF-001", Low, fmt.Sprintf("cold start is %d ms", m), "Aim for under 2 s.")
	}
	a.collectCrashes(ctx, "startup")

	a.step("screen analysis")
	a.rep.Screenshot = a.screenshot(ctx, "main")
	a.uiAudit(ctx)

	_ = a.sh(ctx, "dumpsys", "gfxinfo", opt.Package, "reset")

	if opt.MonkeyEvents > 0 {
		a.step(fmt.Sprintf("monkey (%d events)", opt.MonkeyEvents))
		seed := opt.Seed
		if seed == 0 {
			seed = 42
		}
		out, _ := r.Run(ctx, "shell", "monkey", "-p", opt.Package, "--throttle", "120", "--pct-syskeys", "0", "--pct-appswitch", "2",
			"--monitor-native-crashes", "-s", strconv.FormatInt(seed, 10), "-v", strconv.Itoa(opt.MonkeyEvents))
		m := ParseMonkey(string(out))
		a.rep.Monkey = &m
		if m.Crashed || m.ANR {
			a.c.add("MOB-AND-MONKEY-001", High, "random-input (monkey) test crashed the app", fmt.Sprintf("seed %d after %d events: %s", seed, m.Events, m.Summary))
		} else if m.Aborted {
			a.c.add("MOB-AND-MONKEY-002", Medium, "monkey run aborted before finishing", fmt.Sprintf("%d of %d events injected", m.Events, opt.MonkeyEvents))
		}
		a.collectCrashes(ctx, "monkey test")
	}

	if v, ok := ParseMeminfo(a.sh(ctx, "dumpsys", "meminfo", opt.Package)); ok {
		a.rep.Memory.PSSMB = float64(int(v*10+0.5)) / 10
		switch {
		case v > 400:
			a.c.add("MOB-AND-PERF-002", Medium, fmt.Sprintf("high memory use: %.0f MB PSS", v), "")
		case v > 250:
			a.c.add("MOB-AND-PERF-002", Low, fmt.Sprintf("memory use is %.0f MB PSS", v), "")
		}
	}
	if f, ok := ParseGfxinfo(a.sh(ctx, "dumpsys", "gfxinfo", opt.Package)); ok && f.Total > 0 {
		a.rep.Frames = &f
		switch {
		case f.Total >= 30 && f.JankyPct > 30:
			a.c.add("MOB-AND-PERF-003", Medium, fmt.Sprintf("%.1f%% of frames are janky", f.JankyPct), fmt.Sprintf("%d of %d frames", f.Janky, f.Total))
		case f.Total >= 30 && f.JankyPct > 15:
			a.c.add("MOB-AND-PERF-003", Low, fmt.Sprintf("%.1f%% of frames are janky", f.JankyPct), fmt.Sprintf("%d of %d frames", f.Janky, f.Total))
		}
	}

	scenarios := opt.Scenarios
	if len(scenarios) == 0 {
		scenarios = AllAndroidScenarios
	}
	for _, s := range scenarios {
		a.step("scenario: " + s)
		if res, ok := a.scenario(ctx, s); ok {
			a.rep.Scenarios = append(a.rep.Scenarios, res)
		}
	}

	if n := a.rep.UIControls; n > 0 {
		var unl, small int
		for _, u := range a.rep.UI {
			if u.Rule == "unlabeled-control" {
				unl++
			} else {
				small++
			}
		}
		if unl > 0 {
			a.c.add("MOB-AND-A11Y-001", Medium, "clickable controls without an accessible label", fmt.Sprintf("%d of %d clickable controls on the main screen (TalkBack cannot announce them)", unl, n))
		}
		if small > 0 {
			a.c.add("MOB-AND-A11Y-002", Low, "touch targets smaller than 48x48 dp", fmt.Sprintf("%d control(s) on the main screen", small))
		}
	}
	a.rep.Findings = a.c.sorted()
	return a.rep, nil
}

func lastLine(s string) string {
	lines := strings.Split(strings.TrimSpace(s), "\n")
	return strings.TrimSpace(lines[len(lines)-1])
}

func (a *runAndroid) uiAudit(ctx context.Context) {
	const dump = "/sdcard/qa_ui_dump.xml"
	a.sh(ctx, "uiautomator", "dump", dump)
	out, err := a.r.Run(ctx, "exec-out", "cat", dump)
	a.sh(ctx, "rm", "-f", dump)
	if err != nil || len(out) == 0 {
		return
	}
	issues, n, err := AnalyzeUI(out, a.opt.Package, a.rep.Device.Density)
	if err != nil {
		return
	}
	a.rep.UI, a.rep.UIControls = issues, n
}

func (a *runAndroid) scenario(ctx context.Context, name string) (ScenarioResult, bool) {
	res := ScenarioResult{Name: name}
	var restore func()
	pkg := a.opt.Package
	switch name {
	case "rotate":
		acc, rot := a.sh(ctx, "settings", "get", "system", "accelerometer_rotation"), a.sh(ctx, "settings", "get", "system", "user_rotation")
		a.sh(ctx, "settings", "put", "system", "accelerometer_rotation", "0")
		a.sh(ctx, "settings", "put", "system", "user_rotation", "1")
		restore = func() {
			a.sh(ctx, "settings", "put", "system", "user_rotation", numOr(rot, "0"))
			a.sh(ctx, "settings", "put", "system", "accelerometer_rotation", numOr(acc, "1"))
		}
	case "font_scale":
		orig := a.sh(ctx, "settings", "get", "system", "font_scale")
		a.sh(ctx, "settings", "put", "system", "font_scale", "2.0")
		restore = func() { a.sh(ctx, "settings", "put", "system", "font_scale", numOr(orig, "1.0")) }
	case "dark_mode":
		a.sh(ctx, "cmd", "uimode", "night", "yes")
		restore = func() { a.sh(ctx, "cmd", "uimode", "night", "auto") }
	case "locale_rtl":
		if a.rep.Device.SDK < 33 {
			return res, false
		}
		a.sh(ctx, "cmd", "locale", "set-app-locales", pkg, "--locales", "ar")
		restore = func() { a.sh(ctx, "cmd", "locale", "set-app-locales", pkg) }
	case "background":
		a.sh(ctx, "input", "keyevent", "KEYCODE_HOME")
		a.sleep(2 * time.Second)
	default:
		return res, false
	}
	if restore != nil {
		defer restore()
	}

	if name != "background" {
		a.sh(ctx, "am", "force-stop", pkg)
		a.sleep(time.Second)
	}
	if _, err := a.launch(ctx); err != nil {
		res.Problems = append(res.Problems, "launch failed: "+err.Error())
	}
	a.sleep(2 * time.Second)
	res.Screenshot = a.screenshot(ctx, name)
	if !a.alive(ctx) {
		res.Problems = append(res.Problems, "app process is not running afterwards")
		a.c.add("MOB-AND-SCN-001", High, "app is not running after scenario: "+name, "")
	}
	for _, c := range a.collectCrashes(ctx, "scenario "+name) {
		res.Problems = append(res.Problems, c.Kind+": "+c.Summary)
	}
	res.Passed = len(res.Problems) == 0
	return res, true
}

func numOr(s, def string) string {
	if _, err := strconv.ParseFloat(s, 64); err != nil {
		return def
	}
	return s
}
