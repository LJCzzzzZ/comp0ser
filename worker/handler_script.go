package worker

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"unicode/utf8"

	"comp0ser/prompts"
)

func (w *worker) handleScriptGen(task *Task) error {
	var p GenScriptPlayLoad
	if err := json.Unmarshal(task.Payload, &p); err != nil {
		return err
	}

	prompt, err := w.renderer.System(prompts.Config{
		Subject:  p.Subject,
		Segments: p.Segments,
		MinChars: p.MinChars,
		MaxChars: p.MaxChars,
		Focus:    p.Focus,
		Hook:     p.Hook,
	})
	if err != nil {
		return err
	}

	contents, err := w.llm.GenScript(context.Background(), w.llm.Model, p.RawText, prompt)
	if err != nil {
		return err
	}
	nar := make([]Narration, 0, len(contents))
	for i, content := range contents {
		nar = append(nar, Narration{
			ID:   i,
			Text: content,
		})
	}

	b, err := json.MarshalIndent(nar, "", "  ")
	if err != nil {
		return err
	}

	id, err := w.fs.Add(fmt.Sprintf("%s.txt", p.Subject), bytes.NewReader(b))
	if err != nil {
		return err
	}

	slog.Info("handle genScript ok",
		"file_id", id,
		"subject", p.Subject,
		"segments", p.Segments,
		"minChars", p.MinChars,
		"maxChars", p.MaxChars,
		"rawBytes", len(p.RawText),
		"rawRune", utf8.RuneCountInString(p.RawText),
	)
	return nil
}
