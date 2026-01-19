package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"comp0ser/cmd/llm"
	"comp0ser/config"
)

type Composer struct {
	cfg    *config.Config
	ff     *FFmpeg
	runner *Runner
	apikey string
}

func NewComposer(cfg *config.Config) (*Composer, error) {
	apikey := os.Getenv("GEMINI_API_KEY")

	ff := NewFFmpeg("ffmpeg")

	runner := &Runner{Timeout: cfg.Timeout}
	return &Composer{
		cfg:    cfg,
		ff:     ff,
		runner: runner,
		apikey: apikey,
	}, nil
}

func (c *Composer) Work(ctx context.Context) error {
	paths, err := c.prepareCmdFd()
	if err != nil {
		return err
	}

	client, err := llm.NewGeminiClient(ctx)
	if err != nil {
		return err
	}

	content, err := os.ReadFile(c.cfg.Input)
	if err != nil {
		return err
	}
	slog.Info("Read Input file successful", slog.Any("file_len", len(content)))

	// 1) LLM -> script
	narrations, err := client.GenScript(ctx, c.cfg.TextModel, string(content), llm.ScriptSystemPrompt)
	if err != nil {
		return err
	}
	slog.Info("gen script ok")

	// TODO: Upgrade NewDefaultClient() to http new verison
	ttsCli := llm.NewDefaultClient()

	wavPaths := make([]string, 0, len(narrations))
	for i, narration := range narrations {
		outPath := filepath.Join(paths.OutDir, fmt.Sprintf("narration_%d.wav", i))
		fmt.Println(narration)
		if err := ttsCli.SynthesizeToFile(context.Background(), narration, outPath); err != nil {
			return err
		}

		wavPaths = append(wavPaths, outPath)
	}

	listPath := filepath.Join(paths.OutDir, "concat_wav_list.txt")
	if err := c.ff.WriteConcatWavList(listPath, wavPaths); err != nil {
		return err
	}

	// 3) concat wav -> paths.AudioWav
	// paths.AudioWav: 合并后的单个 wav 输出，比如 /xxx/audio.wav
	concatCmd := c.ff.ConcatWav(listPath, paths.AudioWav)
	if err := c.runner.Run(context.Background(), concatCmd); err != nil {
		return err
	}

	// 4) blend m4a
	blendCmd := c.ff.BlendM4A(paths.AudioWav, c.cfg.BGM, paths.MixM4A, 0.18, true)
	if err := c.runner.Run(context.Background(), blendCmd); err != nil {
		return err
	}

	slog.Info("mix.m4a product successful")

	// 5) render mp4
	renderCmd := c.ff.RenderMp4(c.cfg.Img, paths.MixM4A, paths.OutMP4, 1920, 1080, 30)
	if err := c.runner.Run(context.Background(), renderCmd); err != nil {
		return err
	}
	slog.Info("out.mp4 product successful")
	return nil
}

func (c *Composer) prepareCmdFd() (*Paths, error) {
	if err := os.MkdirAll(c.cfg.Target, 0o755); err != nil {
		return nil, err
	}

	target := c.cfg.Target
	return &Paths{
		OutDir:    target,
		ScriptTxt: filepath.Join(target, "gemini.txt"),
		AudioWav:  filepath.Join(target, "audio.wav"),
		MixM4A:    filepath.Join(target, "mix.m4a"),
		OutMP4:    filepath.Join(target, "target.mp4"),
	}, nil
}
