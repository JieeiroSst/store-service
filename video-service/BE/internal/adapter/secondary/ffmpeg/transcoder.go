package ffmpeg

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/JIeeiroSst/video-service/internal/domain/port"
)

const segmentSeconds = 4

type rendition struct {
	height       int
	videoBitrate int // kbit/s
}

var ladder = []rendition{
	{1080, 5000},
	{720, 2800},
	{480, 1400},
	{360, 800},
}

const audioBitrate = 128 // kbit/s

type Transcoder struct{}

func NewTranscoder() *Transcoder { return &Transcoder{} }

type probeInfo struct {
	width, height int
	duration      float64
	hasAudio      bool
}

func (t *Transcoder) Transcode(ctx context.Context, src, outDir string) (*port.TranscodeResult, error) {
	info, err := probe(ctx, src)
	if err != nil {
		return nil, fmt.Errorf("probe: %w", err)
	}
	rends := pick(info.height)
	for i := range rends {
		if err := os.MkdirAll(filepath.Join(outDir, "v"+strconv.Itoa(i)), 0o755); err != nil {
			return nil, err
		}
	}

	if err := run(ctx, hlsArgs(src, outDir, info, rends)); err != nil {
		return nil, fmt.Errorf("hls: %w", err)
	}
	if err := run(ctx, thumbArgs(src, filepath.Join(outDir, "thumbnail.jpg"), info.duration)); err != nil {
		return nil, fmt.Errorf("thumbnail: %w", err)
	}
	return &port.TranscodeResult{Duration: info.duration, Width: info.width, Height: info.height}, nil
}

func pick(srcHeight int) []rendition {
	var out []rendition
	for _, r := range ladder {
		if r.height <= srcHeight {
			out = append(out, r)
		}
	}
	if len(out) == 0 {
		out = []rendition{{even(srcHeight), ladder[len(ladder)-1].videoBitrate}}
	}
	return out
}

func even(n int) int { return n - n%2 }

func hlsArgs(src, outDir string, info probeInfo, rends []rendition) []string {
	n := len(rends)

	var split, scale []string
	for i, r := range rends {
		split = append(split, fmt.Sprintf("[s%d]", i))
		scale = append(scale, fmt.Sprintf("[s%d]scale=-2:%d[v%d]", i, r.height, i))
	}
	filter := fmt.Sprintf("[0:v:0]split=%d%s;%s", n, strings.Join(split, ""), strings.Join(scale, ";"))

	args := []string{"-y", "-i", src, "-filter_complex", filter}
	var varMap []string
	for i, r := range rends {
		s := strconv.Itoa(i)
		args = append(args,
			"-map", "[v"+s+"]",
			"-c:v:"+s, "libx264", "-preset", "veryfast", "-profile:v:"+s, "main", "-pix_fmt", "yuv420p",
			"-b:v:"+s, fmt.Sprintf("%dk", r.videoBitrate),
			"-maxrate:v:"+s, fmt.Sprintf("%dk", r.videoBitrate*107/100),
			"-bufsize:v:"+s, fmt.Sprintf("%dk", r.videoBitrate*3/2),
		)
		if info.hasAudio {
			args = append(args, "-map", "0:a:0", "-c:a:"+s, "aac", "-b:a:"+s, fmt.Sprintf("%dk", audioBitrate), "-ac", "2")
			varMap = append(varMap, fmt.Sprintf("v:%d,a:%d", i, i))
		} else {
			varMap = append(varMap, fmt.Sprintf("v:%d", i))
		}
	}
	return append(args,
		"-force_key_frames", fmt.Sprintf("expr:gte(t,n_forced*%d)", segmentSeconds),
		"-sc_threshold", "0",
		"-f", "hls",
		"-hls_time", strconv.Itoa(segmentSeconds),
		"-hls_playlist_type", "vod",
		"-hls_flags", "independent_segments",
		"-hls_segment_type", "mpegts",
		"-master_pl_name", "master.m3u8",
		"-hls_segment_filename", filepath.Join(outDir, "v%v", "seg_%04d.ts"),
		"-var_stream_map", strings.Join(varMap, " "),
		filepath.Join(outDir, "v%v", "index.m3u8"),
	)
}

func thumbArgs(src, dst string, duration float64) []string {
	at := min(1.0, duration/2)
	return []string{"-y", "-ss", strconv.FormatFloat(at, 'f', 2, 64), "-i", src,
		"-frames:v", "1", "-vf", "scale=640:-2", "-q:v", "3", dst}
}

func run(ctx context.Context, args []string) error {
	cmd := exec.CommandContext(ctx, "ffmpeg", append([]string{"-hide_banner", "-loglevel", "error", "-nostdin"}, args...)...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if len(msg) > 1000 {
			msg = msg[len(msg)-1000:]
		}
		return fmt.Errorf("%w: %s", err, msg)
	}
	return nil
}

func probe(ctx context.Context, src string) (probeInfo, error) {
	out, err := exec.CommandContext(ctx, "ffprobe", "-v", "error", "-print_format", "json",
		"-show_streams", "-show_format", src).Output()
	if err != nil {
		return probeInfo{}, err
	}
	var p struct {
		Streams []struct {
			CodecType string `json:"codec_type"`
			Width     int    `json:"width"`
			Height    int    `json:"height"`
			Duration  string `json:"duration"`
		} `json:"streams"`
		Format struct {
			Duration string `json:"duration"`
		} `json:"format"`
	}
	if err := json.Unmarshal(out, &p); err != nil {
		return probeInfo{}, err
	}

	var info probeInfo
	for _, s := range p.Streams {
		switch s.CodecType {
		case "video":
			if info.height == 0 {
				info.width, info.height = s.Width, s.Height
				info.duration, _ = strconv.ParseFloat(s.Duration, 64)
			}
		case "audio":
			info.hasAudio = true
		}
	}
	if d, err := strconv.ParseFloat(p.Format.Duration, 64); err == nil && d > 0 {
		info.duration = d
	}
	if info.height == 0 {
		return probeInfo{}, fmt.Errorf("no video stream")
	}
	return info, nil
}
