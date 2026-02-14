package cmd

import (
	"fmt"
	"path/filepath"
	"strings"
)

type Whisper struct {
	Bin   string
	Model string
}

func NewWhisper(bin, model string) *Whisper {
	return &Whisper{
		Bin:   bin,
		Model: model,
	}
}

func (w *Whisper) GenSubtitle(audioPath, outSrtPath, lang string) (*Cmd, error) {
	if audioPath == "" {
		return nil, fmt.Errorf("audioPath is empty")
	}
	if w.Bin == "" || w.Model == "" {
		return nil, fmt.Errorf("whisper bin or model is empty")
	}
	if outSrtPath == "" {
		return nil, fmt.Errorf("outSrtPath is empty")
	}
	if strings.TrimSpace(lang) == "" {
		lang = "auto"
	}

	outPrefix := strings.TrimSuffix(outSrtPath, filepath.Ext(outSrtPath))
	if outPrefix == "" {
		outPrefix = outSrtPath
	}

	args := []string{
		"-m", w.Model,
		"-f", audioPath,
		"-osrt",
		"-l", lang,
		"-of", outPrefix,
	}

	return &Cmd{
		Bin:     w.Bin,
		Args:    args,
		Inputs:  []string{audioPath, w.Model},
		Outputs: []string{outPrefix + ".srt"},
	}, nil
}
