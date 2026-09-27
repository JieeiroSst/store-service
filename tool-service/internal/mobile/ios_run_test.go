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

const simList = `{"devices":{
 "com.apple.CoreSimulator.SimRuntime.iOS-17-5":[
  {"udid":"11111111-1111-1111-1111-111111111111","name":"iPhone 15","state":"Shutdown","isAvailable":true},
  {"udid":"22222222-2222-2222-2222-222222222222","name":"iPhone 15 Pro","state":"Booted","isAvailable":true},
  {"udid":"33333333-3333-3333-3333-333333333333","name":"iPhone SE","state":"Shutdown","isAvailable":false}],
 "com.apple.CoreSimulator.SimRuntime.watchOS-10-5":[{"udid":"44444444-4444-4444-4444-444444444444","name":"Watch","state":"Booted","isAvailable":true}]}}`

const bootedUDID = "22222222-2222-2222-2222-222222222222"

func healthyIOS(t *testing.T, dir string) *fakeADB {
	f := &fakeADB{}
	f.on("list devices available", simList).
		on("get_app_container", "/Users/x/Library/Developer/CoreSimulator/Devices/.../Shop.app\n").
		on("launch "+bootedUDID, "com.acme.shop: 4711\n").
		on("launchctl list", "PID\tStatus\tLabel\n4711\t0\tUIKitApplication:com.acme.shop[0x1a2b][rb-legacy]\n-\t0\tcom.apple.other\n").
		on("ui "+bootedUDID+" appearance", "light\n").
		on("ui "+bootedUDID+" content_size", "large\n").
		on("log show", "Timestamp Thread Type Activity PID TTL\n")
	return f
}

func iosOpts(t *testing.T) IOSOptions {
	return IOSOptions{BundleID: "com.acme.shop", ExecName: "Shop", ArtifactDir: t.TempDir(), CrashDir: t.TempDir(), Sleep: func(time.Duration) {}}
}

func TestParseSimulators(t *testing.T) {
	s, err := ParseSimulators([]byte(simList))
	if err != nil {
		t.Fatal(err)
	}
	if len(s) != 2 || s[0].Name != "iPhone 15 Pro" || s[0].State != "Booted" || s[0].Runtime != "iOS-17-5" {
		t.Fatalf("booted iPhone first, unavailable and watchOS excluded: %+v", s)
	}
	if _, err := ParseSimulators([]byte("not json")); err == nil {
		t.Error("garbage must fail")
	}
}

func TestHealthyIOSApp(t *testing.T) {
	o := iosOpts(t)
	f := healthyIOS(t, o.ArtifactDir)
	rep, err := RunIOS(context.Background(), f, o)
	if err != nil {
		t.Fatal(err)
	}
	if len(rep.Findings) != 0 || rep.PID != "4711" || rep.Device.Name != "iPhone 15 Pro" {
		t.Fatalf("%+v", rep)
	}
	if len(rep.Scenarios) != 4 {
		t.Fatalf("scenarios: %+v", rep.Scenarios)
	}
	for _, want := range []string{"appearance dark", "appearance light", "accessibility-extra-extra-extra-large", "content_size large", "-AppleLanguages (ar) -AppleLocale ar_SA"} {
		if !f.ran(want) {
			t.Errorf("expected %q\n%v", want, f.commands)
		}
	}
	if f.ran("boot " + bootedUDID) {
		t.Error("an already booted simulator must not be booted again")
	}
}

func TestCrashReportsAreFound(t *testing.T) {
	o := iosOpts(t)
	f := healthyIOS(t, o.ArtifactDir)
	report := `{"app_name":"Shop","bug_type":"309"}` + "\n" + `{"exception":{"codes":"0x1","type":"EXC_BAD_ACCESS","signal":"SIGSEGV"},"termination":{"namespace":"SIGNAL","indicator":"Segmentation fault: 11"}}`
	os.WriteFile(filepath.Join(o.CrashDir, "Shop-2026-09-27-101112.ips"), []byte(report), 0o644)
	os.WriteFile(filepath.Join(o.CrashDir, "Other-2026-09-27.ips"), []byte(report), 0o644)
	old := filepath.Join(o.CrashDir, "Shop-2020-01-01-000000.ips")
	os.WriteFile(old, []byte(report), 0o644)
	os.Chtimes(old, time.Now().Add(-72*time.Hour), time.Now().Add(-72*time.Hour))

	rep, err := RunIOS(context.Background(), f, o)
	if err != nil {
		t.Fatal(err)
	}
	if len(rep.Crashes) != 1 || !strings.Contains(rep.Crashes[0].Summary, "EXC_BAD_ACCESS") || !strings.Contains(rep.Crashes[0].Summary, "Segmentation fault") {
		t.Fatalf("only this run's crash of this app counts: %+v", rep.Crashes)
	}
	if ids(rep.Findings)["MOB-IOS-RT-001"] != High {
		t.Fatalf("%+v", rep.Findings)
	}
}

func TestIOSProblems(t *testing.T) {
	ctx := context.Background()
	o := iosOpts(t)
	f := healthyIOS(t, o.ArtifactDir)
	f.rules = append([]rule{{match: "launchctl list", out: "PID\tStatus\tLabel\n-\t9\tUIKitApplication:com.acme.shop[0x1]\n"}}, f.rules...)
	o.DeepLinks = []string{"acme://cart/42"}
	f.rules = append([]rule{{match: "openurl", out: "An error was encountered processing the command (domain=NSOSStatusErrorDomain, code=-10814):\nThe operation couldn't be completed. OSStatus error -10814.\n", err: errors.New("exit 1")}}, f.rules...)
	f.rules = append([]rule{{match: "log show", out: "Timestamp Thread Type\n2026-09-27 10:00:00.1 0x1 Fault 0 0 Shop: (libsystem) bad thing\n2026-09-27 10:00:01.1 0x1 Fault 0 0 Shop: worse thing\n"}}, f.rules...)
	rep, err := RunIOS(ctx, f, o)
	if err != nil {
		t.Fatal(err)
	}
	got := ids(rep.Findings)
	for id, sev := range map[string]string{"MOB-IOS-LAUNCH-002": High, "MOB-IOS-SCN-001": High, "MOB-IOS-DL-001": Medium, "MOB-IOS-LOG-001": Low} {
		if got[id] != sev {
			t.Errorf("%s want %s got %q (%v)", id, sev, got[id], got)
		}
	}
	if rep.LogFaults != 2 {
		t.Errorf("faults=%d", rep.LogFaults)
	}

	launchFail := healthyIOS(t, o.ArtifactDir)
	launchFail.rules = append([]rule{{match: "launch " + bootedUDID, out: "An error was encountered processing the command (domain=FBSOpenApplicationServiceErrorDomain, code=1):\nThe request to open \"com.acme.shop\" failed.\n", err: errors.New("exit 1")}}, launchFail.rules...)
	rep, _ = RunIOS(ctx, launchFail, iosOpts(t))
	if ids(rep.Findings)["MOB-IOS-LAUNCH-001"] != High {
		t.Errorf("launch failure: %+v", rep.Findings)
	}
}

func TestIOSBootAndPreconditions(t *testing.T) {
	ctx := context.Background()
	o := iosOpts(t)
	o.Device = "iPhone 15"
	f := healthyIOS(t, o.ArtifactDir)
	f.rules = append([]rule{
		{match: "launch 11111111", out: "com.acme.shop: 99\n"},
		{match: "launchctl list", out: "99\t0\tUIKitApplication:com.acme.shop[0x1]\n"},
		{match: "ui 11111111", out: "light\n"},
	}, f.rules...)
	o.AppPath = "/tmp/Shop.app"
	rep, err := RunIOS(ctx, f, o)
	if err != nil {
		t.Fatal(err)
	}
	shutdownUDID := "11111111-1111-1111-1111-111111111111"
	if !f.ran("boot "+shutdownUDID) || !f.ran("bootstatus "+shutdownUDID+" -b") || !f.ran("install "+shutdownUDID+" /tmp/Shop.app") || !rep.Installed {
		t.Errorf("shutdown simulator must be booted and the app installed: %v", f.commands)
	}

	for name, bad := range map[string]IOSOptions{
		"bundle":   {BundleID: "bad bundle; rm"},
		"device":   {BundleID: "com.acme.shop", Device: "x; reboot"},
		"exec":     {BundleID: "com.acme.shop", ExecName: `a" OR 1==1 --`},
		"deeplink": {BundleID: "com.acme.shop", DeepLinks: []string{"acme://x; rm -rf /"}},
	} {
		if _, err := RunIOS(ctx, &fakeADB{}, bad); err == nil {
			t.Errorf("%s must be rejected", name)
		}
	}
	none := &fakeADB{}
	none.on("list devices", simList)
	if _, err := RunIOS(ctx, none, IOSOptions{BundleID: "com.acme.shop", Device: "iPad Pro"}); err == nil {
		t.Error("unknown simulator must fail")
	}
	notInstalled := healthyIOS(t, o.ArtifactDir)
	notInstalled.rules = append([]rule{{match: "get_app_container", out: "", err: errors.New("exit 2")}}, notInstalled.rules...)
	if _, err := RunIOS(ctx, notInstalled, iosOpts(t)); err == nil || !strings.Contains(err.Error(), "not installed") {
		t.Errorf("%v", err)
	}
}
