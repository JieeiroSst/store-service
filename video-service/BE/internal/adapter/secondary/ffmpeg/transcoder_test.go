package ffmpeg

import (
	"strings"
	"testing"
)

func TestPickSkipsRenditionsTallerThanSource(t *testing.T) {
	heights := func(rs []rendition) (out []int) {
		for _, r := range rs {
			out = append(out, r.height)
		}
		return
	}
	for src, want := range map[int][]int{
		2160: {1080, 720, 480, 360},
		720:  {720, 480, 360},
		480:  {480, 360},
		240:  {240}, // smaller than every rung: keep source size
		241:  {240}, // and stay even for libx264
	} {
		got := heights(pick(src))
		if len(got) != len(want) {
			t.Fatalf("src %d: got %v want %v", src, got, want)
		}
		for i := range got {
			if got[i] != want[i] {
				t.Fatalf("src %d: got %v want %v", src, got, want)
			}
		}
	}
}

func TestHLSArgsAudioMapping(t *testing.T) {
	rends := pick(720)
	with := strings.Join(hlsArgs("in", "out", probeInfo{hasAudio: true}, rends), " ")
	without := strings.Join(hlsArgs("in", "out", probeInfo{}, rends), " ")
	if !strings.Contains(with, "v:0,a:0 v:1,a:1 v:2,a:2") {
		t.Fatalf("audio var_stream_map missing: %s", with)
	}
	if strings.Contains(without, "a:0") || !strings.Contains(without, "v:0 v:1 v:2") {
		t.Fatalf("silent source must not map audio: %s", without)
	}
}
