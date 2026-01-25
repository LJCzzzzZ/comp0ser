package worker

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
)

func (w *worker) handleTTSAll(task *Task) error {
	var p GenTTSPlayLoad
	if err := json.Unmarshal(task.Payload, &p); err != nil {
		return err
	}

	if p.FileID == "" {
		return fmt.Errorf("empty fileID")
	}

	name, f := w.fs.Get(p.FileID)
	defer f.Close()

	var nars []Narration
	if err := json.NewDecoder(f).Decode(&nars); err != nil {
		return fmt.Errorf("read file %s(%s): %w", p.FileID, name, err)
	}

	for _, nar := range nars {
		b, err := w.tts.Synthesize(context.Background(), nar.Text)
		if err != nil {
			return fmt.Errorf("tts failed idx = %d: %w", nar.ID, err)
		}

		id, err := w.fs.Add(fmt.Sprintf("%04d.wav", nar.ID), bytes.NewReader(b))
		if err != nil {
			return fmt.Errorf("save wav failed: %w", err)
		}

		slog.Info("save wav ok",
			"fildID", id,
		)

	}
	return nil
}
