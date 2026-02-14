package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"unicode/utf8"

	"comp0ser/prompts"
)

func (w *worker) handleScriptGen(task *Task) error {
	var p GenScriptPayLoad
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

	contents, err := w.llm.GenScript(context.Background(), p.Model, p.RawText, prompt)
	if err != nil {
		return err
	}

	doc, err := w.fs.New(p.Subject)

	slog.Info("local file store 'New' ok",
		"doc", doc,
	)
	if err != nil {
		return err
	}

	for i, content := range contents {
		id, err := w.fs.Append(p.Subject, fmt.Sprintf("%04d", i), content, nil)
		if err != nil {
			return err
		}
		slog.Info("append narration ok",
			"idx", i,
			"narID", id,
		)
	}

	slog.Info("handle genScript ok",
		"doc_name", doc,
		"subject", p.Subject,
		"segments", p.Segments,
		"minChars", p.MinChars,
		"maxChars", p.MaxChars,
		"rawBytes", len(p.RawText),
		"rawRune", utf8.RuneCountInString(p.RawText),
	)
	return nil
}
