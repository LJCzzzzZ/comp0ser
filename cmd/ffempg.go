package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type FFmpeg struct {
	Bin string
}

func NewFFmpeg(bin string) *FFmpeg {
	if bin == "" {
		bin = "ffmpeg"
	}
	return &FFmpeg{Bin: bin}
}

func escapeForFFmpegConcatPath(p string) string {
	// concat list 文件里一般是：file '/abs/path'
	// 如果路径里有单引号，按 ffmpeg 的写法转义
	return strings.ReplaceAll(p, "'", "'\\''")
}

// WriteConcatWavList 写出 concat demuxer 的 list 文件
// list 文件内容类似：
// file '/abs/path/seg_001.wav'
// file '/abs/path/seg_002.wav'
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

func (f *FFmpeg) ConcatWav(files, outWav string) *Cmd {
	if outWav == "" {
		outWav = "audio.wav"
	}

	args := []string{
		"-y",
		"-hide_banner",
		"-loglevel", "error",
		"-f", "concat",
		"-safe", "0",
		"-i", files,
		"-ar", fmt.Sprintf("%d", 24000),
		"-ac", "1",
		"-c:a", "pcm_s16le",
		outWav,
	}

	return &Cmd{
		Bin:     f.Bin,
		Args:    args,
		Inputs:  []string{files},
		Outputs: []string{outWav},
	}
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

func (f *FFmpeg) RenderMp4(img, audio, out string, width, hieght, fps int) *Cmd {
	if width <= 0 {
		width = 1920
	}
	if hieght <= 0 {
		hieght = 1080
	}
	if fps <= 0 {
		fps = 30
	}

	if out == "" {
		out = "out.mp4"
	}
	vf := fmt.Sprintf(
		"scale=%d:%d:force_original_aspect_ratio=decrease,pad=%d:%d:(ow-iw)/2:(oh-ih)/2",
		width, hieght, width, hieght,
	)
	args := []string{
		"-y",
		"-loop", "1", "-i", img,
		"-i", audio,
		"-vf", vf,
		"-r", fmt.Sprintf("%d", fps),
		"-c:v", "libx264", "-tune", "stillimage", "-pix_fmt", "yuv420p",
		"-c:a", "aac", "-b:a", "192k",
		"-shortest",
		out,
	}
	return &Cmd{
		Bin:     "ffmpeg",
		Args:    args,
		Inputs:  []string{img, audio},
		Outputs: []string{out},
	}
}
