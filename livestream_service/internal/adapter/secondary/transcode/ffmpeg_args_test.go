package transcode

import (
	"strings"
	"testing"

	"github.com/JIeeiroSst/livestream-service/config"
)

func testTranscodeConfig() config.TranscodeConfig {
	return config.TranscodeConfig{
		SegmentTime:     4,
		SegmentListSize: 6,
	}
}

func TestBuildArgsIncludesABRLadderAndHLSFlags(t *testing.T) {
	args := buildArgs("rtmp://127.0.0.1:1935/live/abc", "/var/hls/abc", testTranscodeConfig())
	joined := strings.Join(args, " ")

	for _, want := range []string{
		"-i rtmp://127.0.0.1:1935/live/abc",
		"split=4[v1][v2][v3][v4]",
		"scale=w=1920:h=1080",
		"scale=w=1280:h=720",
		"scale=w=854:h=480",
		"scale=w=640:h=360",
		"-map [v1out] -c:v:0 libx264 -b:v:0 5000k",
		"-map a:0 -c:a:0 aac -b:a:0 128k",
		"-hls_time 4",
		"-hls_list_size 6",
		"-hls_flags delete_segments+independent_segments",
		"-master_pl_name master.m3u8",
		"-var_stream_map v:0,a:0 v:1,a:1 v:2,a:2 v:3,a:3",
		"/var/hls/abc/v%v_seg%d.ts",
		"/var/hls/abc/v%v.m3u8",
	} {
		if !strings.Contains(joined, want) {
			t.Errorf("buildArgs() missing %q\ngot: %s", want, joined)
		}
	}
}

func TestBuildArgsKeyframeIntervalTracksSegmentTime(t *testing.T) {
	cfg := testTranscodeConfig()
	cfg.SegmentTime = 2
	args := buildArgs("rtmp://in", "/out", cfg)
	joined := strings.Join(args, " ")

	// GOP must equal segmentTime * fps(30) so every rendition's keyframes
	// land on the same segment boundary - otherwise ABR switching glitches.
	if !strings.Contains(joined, "-g 60 -keyint_min 60") {
		t.Errorf("expected -g 60 -keyint_min 60 for a 2s segment, got: %s", joined)
	}
}

func TestBuildArgsEndsWithVariantPlaylistPath(t *testing.T) {
	args := buildArgs("rtmp://in", "/out", testTranscodeConfig())
	last := args[len(args)-1]
	if last != "/out/v%v.m3u8" {
		t.Errorf("expected last arg to be the variant playlist template, got %q", last)
	}
}
