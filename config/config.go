// Package config provide config
package config

import (
	"flag"
	"fmt"
	"time"
)

type Config struct {
	Input string
	BGM   string
	Img   string

	Timeout time.Duration

	Voice  string
	Volume float64

	TextModel string
	TTSModel  string

	LogFile string
	Silent  bool
	Debug   bool

	Target string
}

func (c *Config) Validate() error {
	if c.Voice == "" {
		return fmt.Errorf("voice is required, use -voice <name>")
	}
	if c.BGM == "" {
		return fmt.Errorf("bgm is required, use -bgm <path>")
	}
	if c.Img == "" {
		return fmt.Errorf("img is required, use -img <path>")
	}
	if c.Input == "" {
		return fmt.Errorf("input is required, use -in <path>")
	}
	return nil
}

func (c *Config) Load() error {
	flag.StringVar(&c.Input, "in", "input.txt", "input file")
	flag.StringVar(&c.Voice, "voice", "", "voice name (required)")
	flag.StringVar(&c.BGM, "bgm", "", "bgm path (required)")
	flag.StringVar(&c.Img, "img", "", "image path (required)")
	flag.Float64Var(&c.Volume, "volume", 0.18, "bgm volume")
	flag.DurationVar(&c.Timeout, "timeout", 2*time.Minute, "ffmpeg timeout")
	flag.StringVar(&c.TextModel, "text", "gemini-3-flash-preview", "llm model")
	flag.StringVar(&c.TTSModel, "tts", "gemini-2.5-pro-preview-tts", "tts model")
	flag.StringVar(&c.LogFile, "log", "run.log", "log file path")
	flag.StringVar(&c.Target, "target", "", "output dir (default = voice)")
	flag.BoolVar(&c.Silent, "silent", false, "do not print logs")
	flag.BoolVar(&c.Debug, "debug", false, "enable debug logs")

	flag.Parse()
	if c.Target == "" {
		c.Target = c.Voice
	}

	if err := c.Validate(); err != nil {
		return err
	}
	return nil
}
