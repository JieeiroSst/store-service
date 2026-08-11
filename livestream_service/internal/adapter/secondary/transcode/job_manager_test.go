package transcode

import (
	"testing"
	"time"

	"github.com/JIeeiroSst/livestream-service/config"
)

func testRunner(maxRestarts int, window time.Duration) *ffmpegRunner {
	return &ffmpegRunner{
		jobs: make(map[string]*job),
		cfg: &config.Config{
			Transcode: config.TranscodeConfig{
				MaxRestartsPerWindow: maxRestarts,
				RestartWindow:        window.String(),
			},
		},
	}
}

func TestScheduleRestartAllowsUpToMaxWithinWindow(t *testing.T) {
	r := testRunner(3, time.Minute)
	j := &job{streamKey: "abc"}

	for i := 0; i < 3; i++ {
		if !r.scheduleRestart(j) {
			t.Fatalf("attempt %d: expected restart to be allowed, got denied", i+1)
		}
	}
	if r.scheduleRestart(j) {
		t.Fatal("4th restart within the window should be denied (exceeds MaxRestartsPerWindow=3)")
	}
}

func TestScheduleRestartForgetsAttemptsOutsideWindow(t *testing.T) {
	r := testRunner(1, 10*time.Millisecond)
	j := &job{streamKey: "abc"}

	if !r.scheduleRestart(j) {
		t.Fatal("first restart should be allowed")
	}
	if r.scheduleRestart(j) {
		t.Fatal("second restart within the window should be denied")
	}

	time.Sleep(20 * time.Millisecond)

	if !r.scheduleRestart(j) {
		t.Fatal("restart after the window has passed should be allowed again")
	}
}

func TestStaleSinceOnlyFlagsJobsPastThreshold(t *testing.T) {
	r := testRunner(5, time.Minute)
	fresh := &job{streamKey: "fresh"}
	fresh.lastActivity.Store(time.Now().UnixNano())
	stale := &job{streamKey: "stale"}
	stale.lastActivity.Store(time.Now().Add(-time.Hour).UnixNano())
	starting := &job{streamKey: "starting"} // lastActivity still zero: no segment produced yet

	r.jobs["fresh"] = fresh
	r.jobs["stale"] = stale
	r.jobs["starting"] = starting

	got := r.StaleSince(30 * time.Second)
	if len(got) != 1 || got[0] != "stale" {
		t.Fatalf("StaleSince() = %v, want only [stale]", got)
	}
}
