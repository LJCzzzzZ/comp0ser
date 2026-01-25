package core

import (
	"bufio"
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

type FFmpeg struct {
	Bin          string
	mu           sync.Mutex
	encoderCache map[string]bool
}

func NewFFmpeg(bin string) *FFmpeg {
	if bin == "" {
		bin = "ffmpeg"
	}
	return &FFmpeg{
		Bin:          bin,
		encoderCache: make(map[string]bool),
	}
}

func (f *FFmpeg) SupportsEncoder(ctx context.Context, encoder string) bool {
	f.mu.Lock()
	if v, ok := f.encoderCache[encoder]; ok {
		f.mu.Unlock()
		return v
	}
	f.mu.Unlock()

	cmd := exec.CommandContext(ctx, f.Bin, "-hide_banner", "-encoders")
	out, err := cmd.Output()
	if err != nil {
		f.mu.Lock()
		f.encoderCache[encoder] = false
		f.mu.Unlock()
		return false
	}

	supported := bytes.Contains(out, []byte(encoder))

	f.mu.Lock()
	f.encoderCache[encoder] = supported
	f.mu.Unlock()

	return supported
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

//func (f *FFmpeg) RenderMp4(img, audio, out string, width, height, fps int) *Cmd {
//if width <= 0 { width = 1920 } if hieght <= 0 { hieght = 1080 } if fps <= 0 { fps = 30 } if out == "" { out = "out.mp4" } vf := fmt.Sprintf( "scale=%d:%d:force_original_aspect_ratio=decrease,pad=%d:%d:(ow-iw)/2:(oh-ih)/2", width, hieght, width, hieght, ) args := []string{ "-y", "-loop", "1", "-i", img, "-i", audio, "-vf", vf, "-r", fmt.Sprintf("%d", fps), "-c:v", "libx264", "-tune", "stillimage", "-pix_fmt", "yuv420p", "-c:a", "aac", "-b:a", "192k", "-shortest", out, } return &Cmd{ Bin: "ffmpeg", Args: args, Inputs: []string{img, audio}, Outputs: []string{out}, }
//}

func writeWav(filename string, pcm []byte, channels uint16, sampleRate uint32, bitsPerSample uint16) error {
	// WAV/RIFF header sizes
	blockAlign := channels * (bitsPerSample / 8)
	byteRate := sampleRate * uint32(blockAlign)
	dataSize := uint32(len(pcm))
	riffSize := 36 + dataSize // 4 + (8 + SubChunk1) + (8 + SubChunk2) = 36 + data

	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	// RIFF header
	if _, err := f.Write([]byte("RIFF")); err != nil {
		return err
	}
	if err := binary.Write(f, binary.LittleEndian, riffSize); err != nil {
		return err
	}
	if _, err := f.Write([]byte("WAVE")); err != nil {
		return err
	}

	// fmt chunk
	if _, err := f.Write([]byte("fmt ")); err != nil {
		return err
	}
	if err := binary.Write(f, binary.LittleEndian, uint32(16)); err != nil { // PCM fmt chunk size
		return err
	}
	if err := binary.Write(f, binary.LittleEndian, uint16(1)); err != nil { // AudioFormat 1 = PCM
		return err
	}
	if err := binary.Write(f, binary.LittleEndian, channels); err != nil {
		return err
	}
	if err := binary.Write(f, binary.LittleEndian, sampleRate); err != nil {
		return err
	}
	if err := binary.Write(f, binary.LittleEndian, byteRate); err != nil {
		return err
	}
	if err := binary.Write(f, binary.LittleEndian, blockAlign); err != nil {
		return err
	}
	if err := binary.Write(f, binary.LittleEndian, bitsPerSample); err != nil {
		return err
	}

	// data chunk
	if _, err := f.Write([]byte("data")); err != nil {
		return err
	}
	if err := binary.Write(f, binary.LittleEndian, dataSize); err != nil {
		return err
	}
	if _, err := f.Write(pcm); err != nil {
		return err
	}

	return nil
}
