// Package config provide config
package config

import (
	"errors"
	"flag"
	"time"
)

var ErrHelp = errors.New("help requested")

type Config struct {
	Input string
	BGM   string
	Img   string

	Timeout time.Duration

	Voice  string
	Volume float64

	TextModel string
	TTSModel  string

	LogFile   string
	LogLevel  string
	LogFormat string
	Silent    bool
	Debug     bool

	HTTPAddr string

	Target string
}

func (c *Config) Validate() error {
	//if c.Voice == "" {
	//return fmt.Errorf("voice is required, use -voice <name>")
	//}
	//if c.BGM == "" {
	//return fmt.Errorf("bgm is required, use -bgm <path>")
	//}
	//if c.Img == "" {
	//return fmt.Errorf("img is required, use -img <path>")
	//}
	//if c.Input == "" {
	//return fmt.Errorf("input is required, use -in <path>")
	//}
	return nil
}

func Load(args []string) (*Config, error) {
	fs := flag.NewFlagSet("comp0ser", flag.ContinueOnError)
	var c Config
	fs.StringVar(&c.HTTPAddr, "addr", ":8088", "http addr")
	fs.StringVar(&c.Input, "in", "input.txt", "input file")
	fs.StringVar(&c.Voice, "voice", "", "voice name (required)")
	fs.StringVar(&c.BGM, "bgm", "", "bgm path (required)")
	fs.StringVar(&c.Img, "img", "", "image path (required)")
	fs.Float64Var(&c.Volume, "volume", 0.18, "bgm volume")
	fs.DurationVar(&c.Timeout, "timeout", 30*time.Minute, "ffmpeg timeout")
	fs.StringVar(&c.TextModel, "text", "gemini-3-flash-preview", "llm model")
	fs.StringVar(&c.TTSModel, "tts", "gemini-2.5-pro-preview-tts", "tts model")
	fs.StringVar(&c.LogFile, "log", "run.log", "log file path")
	fs.StringVar(&c.Target, "target", "", "output dir (default = voice)")
	fs.BoolVar(&c.Silent, "silent", false, "do not print logs")
	fs.BoolVar(&c.Debug, "debug", false, "enable debug logs")

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil, ErrHelp
		}
		return nil, err
	}

	if c.Target == "" {
		c.Target = c.Voice
	}

	if err := c.Validate(); err != nil {
		return nil, err
	}

	return &c, nil
}

