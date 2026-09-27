package mobile

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const pngHeader = "\x89PNG\r\n\x1a\nfake"

type fakeADB struct {
	rules    []rule
	commands []string
}

type rule struct {
	match string
	out   string
	err   error
	once  bool
	used  bool
}

func (f *fakeADB) on(match, out string) *fakeADB {
	f.rules = append(f.rules, rule{match: match, out: out})
	return f
}

func (f *fakeADB) Run(ctx context.Context, args ...string) ([]byte, error) {
	cmd := strings.Join(args, " ")
	f.commands = append(f.commands, cmd)
	for i := range f.rules {
		r := &f.rules[i]
		if strings.Contains(cmd, r.match) {
			return []byte(r.out), r.err
		}
	}
	return nil, nil
}

func (f *fakeADB) ran(sub string) bool {
	for _, c := range f.commands {
		if strings.Contains(c, sub) {
			return true
		}
	}
	return false
}

const amStart = "Starting: Intent { cmp=com.acme.app/.MainActivity }\nStatus: ok\nLaunchState: COLD\nActivity: com.acme.app/.MainActivity\nTotalTime: %d\nWaitTime: 900\nComplete\n"

func healthyDevice(startMs string) *fakeADB {
	f := &fakeADB{}
	f.on("getprop sys.boot_completed", "1\n").
		on("ro.product.model", "Pixel 8\n").on("ro.build.version.release", "14\n").on("ro.build.version.sdk", "34\n").
		on("wm size", "Physical size: 1080x2400\n").on("wm density", "Physical density: 420\n").
		on("pm list packages", "package:com.acme.app\n").
		on("resolve-activity", "priority=0 preferredOrder=0 match=0x108000 specificIndex=-1 isDefault=false\ncom.acme.app/.MainActivity\n").
		on("am start -W", strings.Replace(amStart, "%d", startMs, 1)).
		on("meminfo", "         TOTAL PSS:   131072            TOTAL RSS:   210000\n").
		on("gfxinfo", "Total frames rendered: 200\nJanky frames: 10 (5.00%)\n90th percentile: 12ms\n99th percentile: 30ms\n").
		on("monkey", ":Monkey: seed=42 count=300\nEvents injected: 300\n// Monkey finished\n").
		on("pidof", "4321\n").
		on("exec-out screencap", pngHeader).
		on("exec-out cat", `<?xml version='1.0' encoding='UTF-8'?><hierarchy rotation="0"><node index="0" text="" class="android.widget.FrameLayout" package="com.acme.app" clickable="false" enabled="true" bounds="[0,0][1080,2400]">
<node index="0" text="Buy" resource-id="com.acme.app:id/buy" class="android.widget.Button" package="com.acme.app" content-desc="" clickable="true" enabled="true" bounds="[100,200][500,400]"/></node></hierarchy>`).
		on("logcat -d", "09-27 10:00:00.000  1000  1000 I ActivityManager: Start proc com.acme.app\n")
	return f
}

func opts(t *testing.T) AndroidOptions {
	return AndroidOptions{Package: "com.acme.app", MonkeyEvents: 300, ArtifactDir: t.TempDir(), Sleep: func(time.Duration) {}}
}

func TestHealthyApp(t *testing.T) {
	f := healthyDevice("812")
	o := opts(t)
	rep, err := RunAndroid(context.Background(), f, o)
	if err != nil {
		t.Fatal(err)
	}
	if len(rep.Findings) != 0 {
		t.Fatalf("healthy app flagged: %+v", rep.Findings)
	}
	if rep.Startup.MedianMs != 812 || len(rep.Startup.ColdMs) != 3 || rep.Memory.PSSMB != 128 || rep.Frames.JankyPct != 5 || rep.Monkey.Events != 300 {
		t.Fatalf("metrics: %+v", rep)
	}
	if rep.Device.Model != "Pixel 8" || rep.Device.SDK != 34 || rep.Device.Density != 420 {
		t.Fatalf("device: %+v", rep.Device)
	}
	if len(rep.Scenarios) != 5 {
		t.Fatalf("scenarios: %+v", rep.Scenarios)
	}
	for _, s := range rep.Scenarios {
		if !s.Passed {
			t.Errorf("scenario %s failed: %v", s.Name, s.Problems)
		}
	}
	if rep.Screenshot == "" {
		t.Fatal("no screenshot")
	}
	if b, _ := os.ReadFile(filepath.Join(o.ArtifactDir, rep.Screenshot)); !strings.HasPrefix(string(b), "\x89PNG") {
		t.Fatal("screenshot not written")
	}
	for _, want := range []string{"settings put system font_scale 2.0", "settings put system font_scale 1.0", "user_rotation 1", "night yes", "night auto", "set-app-locales com.acme.app --locales ar", "KEYCODE_HOME"} {
		if !f.ran(want) {
			t.Errorf("expected command %q", want)
		}
	}
}

func TestDeviceStateIsRestored(t *testing.T) {
	f := healthyDevice("500")
	f.rules = append([]rule{{match: "settings get system font_scale", out: "1.3\n"}, {match: "settings get system user_rotation", out: "0\n"}, {match: "settings get system accelerometer_rotation", out: "1\n"}}, f.rules...)
	if _, err := RunAndroid(context.Background(), f, opts(t)); err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"font_scale 1.3", "user_rotation 0", "accelerometer_rotation 1", "set-app-locales com.acme.app"} {
		if !f.ran(want) {
			t.Errorf("original setting not restored: %q\n%v", want, f.commands)
		}
	}
}

const crashLog = `09-27 10:11:12.123  4321  4321 E AndroidRuntime: FATAL EXCEPTION: main
09-27 10:11:12.123  4321  4321 E AndroidRuntime: Process: com.acme.app, PID: 4321
09-27 10:11:12.123  4321  4321 E AndroidRuntime: java.lang.NullPointerException: Attempt to invoke virtual method on a null object reference
09-27 10:11:12.123  4321  4321 E AndroidRuntime: 	at com.acme.app.Cart.total(Cart.java:42)
09-27 10:11:12.123  4321  4321 E AndroidRuntime: 	at com.acme.app.Main.onClick(Main.java:10)
09-27 10:11:13.000   900   900 I ActivityManager: Process com.acme.app (pid 4321) has died
09-27 10:11:14.000  5555  5555 E AndroidRuntime: FATAL EXCEPTION: main
09-27 10:11:14.000  5555  5555 E AndroidRuntime: Process: com.other.app, PID: 5555
09-27 10:11:14.000  5555  5555 E AndroidRuntime: java.lang.IllegalStateException: not ours
09-27 10:12:00.000   900   950 E ActivityManager: ANR in com.acme.app (com.acme.app/.MainActivity)
09-27 10:12:00.000   900   950 E ActivityManager: Reason: Input dispatching timed out
09-27 10:13:00.000     0     0 F libc    : Fatal signal 11 (SIGSEGV), code 1 (SEGV_MAPERR), fault addr 0x0 in tid 4400 (RenderThread), pid 4321 (com.acme.app)
09-27 10:13:01.000  4321  4321 W System   : StrictMode policy violation; ~duration=12 ms
`

func TestCrashesAreDetectedAndAttributed(t *testing.T) {
	cs, strict := ParseCrashes(crashLog, "com.acme.app")
	if len(cs) != 3 || strict != 1 {
		t.Fatalf("crashes=%+v strict=%d", cs, strict)
	}
	if cs[0].Kind != "java" || !strings.Contains(cs[0].Summary, "NullPointerException") || len(cs[0].Stack) != 2 || !strings.Contains(cs[0].Stack[0], "Cart.java:42") {
		t.Errorf("java crash: %+v", cs[0])
	}
	if cs[1].Kind != "anr" || !strings.Contains(cs[1].Summary, "Input dispatching timed out") {
		t.Errorf("anr: %+v", cs[1])
	}
	if cs[2].Kind != "native" || !strings.Contains(cs[2].Summary, "SIGSEGV") {
		t.Errorf("native: %+v", cs[2])
	}
	if other, _ := ParseCrashes(crashLog, "com.other.app"); len(other) != 1 || !strings.Contains(other[0].Summary, "not ours") {
		t.Errorf("other app's crash must be attributed to it only: %+v", other)
	}
	if none, _ := ParseCrashes(crashLog, "com.acme"); len(none) != 0 {
		t.Errorf("package prefix must not match: %+v", none)
	}
}

func TestCrashingAppFailsRun(t *testing.T) {
	f := healthyDevice("812")
	f.rules = append([]rule{
		{match: "logcat -d", out: crashLog},
		{match: "monkey", out: ":Monkey: seed=42 count=300\n// CRASH: com.acme.app (pid 4321)\n// Short Msg: java.lang.NullPointerException\n** Monkey aborted due to error.\nEvents injected: 87\n"},
	}, f.rules...)
	rep, err := RunAndroid(context.Background(), f, opts(t))
	if err != nil {
		t.Fatal(err)
	}
	got := ids(rep.Findings)
	for id, sev := range map[string]string{"MOB-AND-RT-001": High, "MOB-AND-RT-002": High, "MOB-AND-RT-003": High, "MOB-AND-MONKEY-001": High, "MOB-AND-STRICT-001": Low} {
		if got[id] != sev {
			t.Errorf("%s want %s got %q (%v)", id, sev, got[id], got)
		}
	}
	if rep.Monkey == nil || !rep.Monkey.Crashed || rep.Monkey.Events != 87 || !rep.Monkey.Aborted {
		t.Errorf("monkey: %+v", rep.Monkey)
	}
	if rep.Findings[0].Severity != High {
		t.Error("findings must be sorted by severity")
	}
}

func TestPerformanceAndAccessibilityFindings(t *testing.T) {
	f := healthyDevice("6400")
	f.rules = append([]rule{
		{match: "meminfo", out: "TOTAL PSS: 524288 TOTAL RSS: 1\n"},
		{match: "gfxinfo", out: "Total frames rendered: 300\nJanky frames: 120 (40.00%)\n"},
		{match: "exec-out cat", out: `<hierarchy><node text="" class="android.widget.ImageButton" resource-id="com.acme.app:id/fav" package="com.acme.app" content-desc="" clickable="true" enabled="true" bounds="[0,0][100,100]"/>
<node text="" class="android.view.ViewGroup" package="com.acme.app" clickable="true" enabled="true" bounds="[0,200][1080,600]"><node text="Open" class="android.widget.TextView" package="com.acme.app" clickable="false" enabled="true" bounds="[0,200][300,300]"/></node>
<node text="" class="android.widget.ImageButton" package="com.android.systemui" clickable="true" enabled="true" bounds="[0,0][10,10]"/></hierarchy>`},
	}, f.rules...)
	rep, err := RunAndroid(context.Background(), f, opts(t))
	if err != nil {
		t.Fatal(err)
	}
	got := ids(rep.Findings)
	for id, sev := range map[string]string{"MOB-AND-PERF-001": Medium, "MOB-AND-PERF-002": Medium, "MOB-AND-PERF-003": Medium, "MOB-AND-A11Y-001": Medium, "MOB-AND-A11Y-002": Low} {
		if got[id] != sev {
			t.Errorf("%s want %s got %q (%v)", id, sev, got[id], got)
		}
	}
	if rep.UIControls != 2 {
		t.Errorf("only the app's own clickable controls count (system UI ignored): %d", rep.UIControls)
	}
	var unlabeled, small int
	for _, u := range rep.UI {
		switch u.Rule {
		case "unlabeled-control":
			unlabeled++
		case "small-touch-target":
			small++
		}
	}
	if unlabeled != 1 || small != 1 {
		t.Errorf("a control whose child has text is labelled; 100px at 420dpi is 38dp: unlabeled=%d small=%d %+v", unlabeled, small, rep.UI)
	}
}

func TestPreconditionsAndValidation(t *testing.T) {
	ctx := context.Background()
	for name, o := range map[string]AndroidOptions{
		"bad package":   {Package: "com.acme; rm -rf /"},
		"bad component": {Package: "com.acme.app", Component: "com.acme.app/.Main; reboot"},
		"bad monkey":    {Package: "com.acme.app", MonkeyEvents: 999999},
	} {
		if _, err := RunAndroid(ctx, &fakeADB{}, o); err == nil {
			t.Errorf("%s must be rejected", name)
		}
	}
	if _, err := RunAndroid(ctx, (&fakeADB{}).on("boot_completed", "\n"), opts(t)); err == nil || !strings.Contains(err.Error(), "not ready") {
		t.Errorf("device not booted: %v", err)
	}
	notInstalled := healthyDevice("1")
	notInstalled.rules = append([]rule{{match: "pm list packages", out: ""}}, notInstalled.rules...)
	if _, err := RunAndroid(ctx, notInstalled, opts(t)); err == nil || !strings.Contains(err.Error(), "not installed") {
		t.Errorf("missing package: %v", err)
	}
	bad := healthyDevice("1")
	bad.rules = append([]rule{{match: "install -r", out: "Performing Streamed Install\nadb: failed to install a.apk: Failure [INSTALL_FAILED_OLDER_SDK]\n", err: errors.New("exit 1")}}, bad.rules...)
	o := opts(t)
	o.APKPath = "/tmp/a.apk"
	if _, err := RunAndroid(ctx, bad, o); err == nil || !strings.Contains(err.Error(), "INSTALL_FAILED_OLDER_SDK") {
		t.Errorf("install failure must be reported: %v", err)
	}
}

func TestLaunchFailureIsAFinding(t *testing.T) {
	f := healthyDevice("1")
	f.rules = append([]rule{{match: "am start -W", out: "Starting: Intent { cmp=com.acme.app/.Gone }\nError type 3\nError: Activity class {com.acme.app/com.acme.app.Gone} does not exist.\n"}}, f.rules...)
	rep, err := RunAndroid(context.Background(), f, opts(t))
	if err != nil {
		t.Fatal(err)
	}
	if ids(rep.Findings)["MOB-AND-LAUNCH-001"] != High {
		t.Fatalf("%+v", rep.Findings)
	}
}

func TestOnlyExpectedAdbCommandsAreIssued(t *testing.T) {
	f := healthyDevice("500")
	o := opts(t)
	o.APKPath = "/tmp/app.apk"
	f.rules = append([]rule{{match: "install -r", out: "Success\n"}}, f.rules...)
	o.Uninstall = true
	if _, err := RunAndroid(context.Background(), f, o); err != nil {
		t.Fatal(err)
	}
	allowed := map[string]bool{"shell": true, "install": true, "uninstall": true, "logcat": true, "exec-out": true}
	for _, c := range f.commands {
		if !allowed[strings.Fields(c)[0]] {
			t.Errorf("unexpected adb subcommand: %q", c)
		}
	}
	if !f.ran("install -r -g /tmp/app.apk") || !f.ran("uninstall com.acme.app") {
		t.Error("install/uninstall not issued")
	}
}

func TestParsers(t *testing.T) {
	devs := ParseDevices("List of devices attached\nemulator-5554          device product:sdk_gphone64 model:sdk_gphone64_arm64 device:emu64a transport_id:1\nR58M123ABC             unauthorized transport_id:2\n\n")
	if len(devs) != 2 || devs[0].Model != "sdk_gphone64_arm64" || devs[1].State != "unauthorized" {
		t.Errorf("%+v", devs)
	}
	if ms, err := ParseStartTime("Status: ok\nTotalTime: 431\n"); err != nil || ms != 431 {
		t.Errorf("%d %v", ms, err)
	}
	if _, err := ParseStartTime("Status: timeout\n"); err == nil {
		t.Error("timeout must be an error")
	}
	if mb, ok := ParseMeminfo("Applications Memory Usage (in Kilobytes):\n           TOTAL   98304   77000  10  0  120000\n"); !ok || mb != 96 {
		t.Errorf("legacy meminfo: %v %v", mb, ok)
	}
	if median([]int{900, 100, 500}) != 500 || median(nil) != 0 {
		t.Error("median")
	}
	if !ValidSerial("emulator-5554") || !ValidSerial("192.168.1.5:5555") || ValidSerial("a b") || ValidSerial("x;y") {
		t.Error("serial validation")
	}
}
