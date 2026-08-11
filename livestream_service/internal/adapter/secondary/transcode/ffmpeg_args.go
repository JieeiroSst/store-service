package transcode

import (
	"fmt"
	"path/filepath"

	"github.com/JIeeiroSst/livestream-service/config"
)

// rendition is one entry in the ABR ladder: 1080p/720p/480p/360p, each
// with its own video+audio bitrate ceiling.
type rendition struct {
	width, height              int
	videoBitrate, maxrate, buf string
	audioBitrate               string
}

var ladder = []rendition{
	{1920, 1080, "5000k", "5350k", "7500k", "128k"},
	{1280, 720, "2800k", "2996k", "4200k", "128k"},
	{854, 480, "1400k", "1498k", "2100k", "96k"},
	{640, 360, "800k", "856k", "1200k", "64k"},
}

// buildArgs constructs the ffmpeg argv for one-decode-many-outputs ABR
// HLS packaging: split the input once, scale into N renditions, and mux
// each into its own HLS variant behind a generated master.m3u8. Mirrors
// the reference shell script this was ported from.
func buildArgs(rtmpInput, outDir string, cfg config.TranscodeConfig) []string {
	args := []string{
		"-hide_banner", "-loglevel", "warning",
		"-fflags", "nobuffer", "-flags", "low_delay",
		"-i", rtmpInput,
		"-filter_complex", filterComplex(),
	}

	for i, r := range ladder {
		args = append(args,
			"-map", fmt.Sprintf("[v%dout]", i+1),
			fmt.Sprintf("-c:v:%d", i), "libx264",
			fmt.Sprintf("-b:v:%d", i), r.videoBitrate,
			fmt.Sprintf("-maxrate:v:%d", i), r.maxrate,
			fmt.Sprintf("-bufsize:v:%d", i), r.buf,
		)
	}
	for i, r := range ladder {
		args = append(args,
			"-map", "a:0",
			fmt.Sprintf("-c:a:%d", i), "aac",
			fmt.Sprintf("-b:a:%d", i), r.audioBitrate,
			"-ac", "2",
		)
	}

	// Fixed GOP aligned to the segment duration so every rendition's
	// keyframes land on the same segment boundary - required for clean
	// ABR switching and correct CDN caching of segments.
	keyint := cfg.SegmentTime * 30
	args = append(args,
		"-preset", "veryfast", "-profile:v", "high", "-sc_threshold", "0",
		"-g", fmt.Sprintf("%d", keyint), "-keyint_min", fmt.Sprintf("%d", keyint),
		"-f", "hls",
		"-hls_time", fmt.Sprintf("%d", cfg.SegmentTime),
		"-hls_list_size", fmt.Sprintf("%d", cfg.SegmentListSize),
		"-hls_flags", "delete_segments+independent_segments",
		"-hls_segment_type", "mpegts",
		"-master_pl_name", "master.m3u8",
		"-hls_segment_filename", filepath.Join(outDir, "v%v_seg%d.ts"),
		"-var_stream_map", "v:0,a:0 v:1,a:1 v:2,a:2 v:3,a:3",
		filepath.Join(outDir, "v%v.m3u8"),
	)
	return args
}

func filterComplex() string {
	return "[0:v]split=4[v1][v2][v3][v4];" +
		fmt.Sprintf("[v1]scale=w=%d:h=%d[v1out];", ladder[0].width, ladder[0].height) +
		fmt.Sprintf("[v2]scale=w=%d:h=%d[v2out];", ladder[1].width, ladder[1].height) +
		fmt.Sprintf("[v3]scale=w=%d:h=%d[v3out];", ladder[2].width, ladder[2].height) +
		fmt.Sprintf("[v4]scale=w=%d:h=%d[v4out]", ladder[3].width, ladder[3].height)
}
