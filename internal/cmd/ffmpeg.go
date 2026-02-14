package cmd

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

type FFmpeg struct {
	Bin string
	mu  sync.Mutex
}

func NewFFmpeg(bin string) *FFmpeg {
	if bin == "" {
		bin = "ffmpeg"
	}
	return &FFmpeg{
		Bin: bin,
	}
}

func escapeForFFmpegConcatPath(p string) string {
	return strings.ReplaceAll(p, "'", "'\\''")
}

func (f *FFmpeg) WriteConcatWavList(listPath string, wavPaths []string) error {
	if len(wavPaths) == 0 {
		return fmt.Errorf("wavPaths is empty")
	}

	if err := os.MkdirAll(filepath.Dir(listPath), 0o755); err != nil {
		return err
	}

	file, err := os.Create(listPath)
	if err != nil {
		return err
	}
	defer file.Close()

	w := bufio.NewWriter(file)
	for _, p := range wavPaths {
		abs, err := filepath.Abs(p)
		if err != nil {
			return err
		}
		abs = escapeForFFmpegConcatPath(abs)

		if _, err := fmt.Fprintf(w, "file '%s'\n", abs); err != nil {
			return err
		}
	}
	return w.Flush()
}

func (f *FFmpeg) ConcatWav(wavs []string, dir, filename string) (*Cmd, error) {
	if len(wavs) == 0 {
		return nil, fmt.Errorf("wavPaths is empty")
	}
	if filename == "" {
		filename = "audio.wav"
	}

	list := filepath.Join(dir, "concat.txt")

	if err := f.WriteConcatWavList(list, wavs); err != nil {
		return nil, err
	}

	args := []string{
		"-y",
		"-hide_banner",
		"-loglevel", "error",
		"-f", "concat",
		"-safe", "0",
		"-i", list,
		"-c:a", "copy",
		filepath.Join(dir, filename),
	}

	return &Cmd{
		Bin:     f.Bin,
		Args:    args,
		Inputs:  wavs,
		Outputs: []string{dir},
	}, nil
}

func (f *FFmpeg) BlendM4A(audio, bgm, out string, volume float64, loop bool) *Cmd {
	vol := volume
	if vol <= 0 {
		vol = 0.18
	}
	if out == "" {
		out = "out.m4a"
	}

	filter := fmt.Sprintf(
		"[1:a]volume=%.3f[a1];[0:a][a1]amix=inputs=2:duration=first:dropout_transition=2[aout]",
		vol,
	)

	args := []string{"-y", "-i", audio}

	if loop {
		args = append(args, "-stream_loop", "-1")
	}
	args = append(args,
		"-i", bgm,
		"-filter_complex", filter,
		"-map", "[aout]",
		"-c:a", "aac",
		"-b:a", "192k",
		out,
	)

	return &Cmd{
		Bin:     "ffmpeg",
		Args:    args,
		Inputs:  []string{audio, bgm},
		Outputs: []string{out},
	}
}

func (f *FFmpeg) BurnSubtitle(videoPath, subPath, outPath string) (*Cmd, error) {
	if videoPath == "" || subPath == "" {
		return nil, fmt.Errorf("videoPath or subPath is empty")
	}
	if outPath == "" {
		outPath = "final_with_sub.mp4"
	}

	args := []string{
		"-y",
		"-i", videoPath,

		"-vf", fmt.Sprintf(
			"subtitles=%s:force_style='FontName=Arial,FontSize=18,Outline=2'",
			subPath,
		),

		"-c:v", "libx264",
		"-preset", "veryfast",
		"-crf", "20",

		"-c:a", "copy",
		"-movflags", "+faststart",

		outPath,
	}

	return &Cmd{
		Bin:     "ffmpeg",
		Args:    args,
		Inputs:  []string{videoPath, subPath},
		Outputs: []string{outPath},
	}, nil
}

func (f *FFmpeg) Merge(videoPath, audioPath, outPath string) (*Cmd, error) {
	if videoPath == "" {
		return nil, fmt.Errorf("videoPath is empty")
	}
	if audioPath == "" {
		return nil, fmt.Errorf("audioPath is empty")
	}
	if outPath == "" {
		outPath = "final.mp4"
	}

	args := []string{
		"-y",

		"-i", videoPath,
		"-i", audioPath,

		"-map", "0:v:0",
		"-map", "1:a:0",

		"-c:v", "copy",
		"-c:a", "copy",

		"-shortest",

		"-movflags", "+faststart",

		outPath,
	}

	return &Cmd{
		Bin:     "ffmpeg",
		Args:    args,
		Inputs:  []string{videoPath, audioPath},
		Outputs: []string{outPath},
	}, nil
}

type seg struct {
	Path        string
	EffectiveTo float64 // effective duration（已减 tailCut）
}

func (f *FFmpeg) ConcatAssets(
	videos []string,
	out string,
	dur float64,
	tailCut float64,
	loop bool,
) (*Cmd, error) {
	if len(videos) == 0 {
		return nil, fmt.Errorf("videos is empty")
	}
	if dur <= 0 {
		return nil, fmt.Errorf("dur must be > 0")
	}
	if tailCut < 0 {
		tailCut = 0
	}
	if out == "" {
		out = "out.mp4"
	}

	fmt.Println(videos)
	fmt.Println("hello")

	items := make([]seg, 0, len(videos))
	for _, v := range videos {
		d, err := probeDurationSeconds(v)
		if err != nil {
			return nil, fmt.Errorf("ffprobe failed: %s: %w", v, err)
		}

		eff := d - tailCut
		if eff < 0.05 {
			eff = 0.05
		}

		items = append(items, seg{
			Path:        v,
			EffectiveTo: eff,
		})
	}

	var seq []seg
	var sum float64

	appendOnce := func() {
		for _, it := range items {
			if sum >= dur {
				return
			}
			seq = append(seq, it)
			sum += it.EffectiveTo
		}
	}

	appendOnce()
	if loop {
		for sum < dur {
			appendOnce()
		}
	}

	if len(seq) == 0 {
		return nil, fmt.Errorf("empty concat sequence")
	}

	args := []string{"-y"}
	for _, s := range seq {
		args = append(args, "-i", s.Path)
	}

	var fc strings.Builder

	for i, s := range seq {
		fmt.Fprintf(
			&fc,
			"[%d:v]"+
				"trim=0:%.3f,"+
				"setpts=PTS-STARTPTS,"+
				"scale=1920:1080,"+
				"setsar=1,"+
				"fps=30"+
				"[v%d];",
			i, s.EffectiveTo, i,
		)
	}

	for i := range seq {
		fmt.Fprintf(&fc, "[v%d]", i)
	}

	fmt.Fprintf(
		&fc,
		"concat=n=%d:v=1:a=0,format=yuv420p[vout]",
		len(seq),
	)

	args = append(args,
		"-filter_complex", fc.String(),
		"-map", "[vout]",
		"-an",

		"-r", "30",
		"-t", fmt.Sprintf("%.3f", dur),

		"-c:v", "libx264",
		"-preset", "veryfast",
		"-crf", "20",
		"-movflags", "+faststart",

		out,
	)

	return &Cmd{
		Bin:     "ffmpeg",
		Args:    args,
		Inputs:  videos,
		Outputs: []string{out},
	}, nil
}

func probeDurationSeconds(path string) (float64, error) {
	cmd := exec.Command("ffprobe",
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=nk=1:nw=1",
		path,
	)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		return 0, fmt.Errorf("ffprobe error: %v, output: %s", err, out.String())
	}
	s := strings.TrimSpace(out.String())
	sec, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, fmt.Errorf("parse duration %q: %w", s, err)
	}
	return sec, nil
}

func writeConcatListWithOutpoint(seq []seg) (string, error) {
	f, err := os.CreateTemp("", "ffconcat-*.txt")
	if err != nil {
		return "", err
	}
	defer f.Close()

	for _, s := range seq {
		if _, err := fmt.Fprintf(f, "file %s\n", ffconcatQuote(s.Path)); err != nil {
			return "", err
		}
		if _, err := fmt.Fprintf(f, "outpoint %.3f\n", s.EffectiveTo); err != nil {
			return "", err
		}
	}
	return f.Name(), nil
}

func ffconcatQuote(p string) string {
	p = filepath.Clean(p)
	p = strings.ReplaceAll(p, "\\", "\\\\")
	p = strings.ReplaceAll(p, "'", "\\'")
	return "'" + p + "'"
}
