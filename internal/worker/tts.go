package worker

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
)

func (w *worker) handleTTSAll(task *Task) error {
	var p GenTTSPayLoad
	if err := json.Unmarshal(task.Payload, &p); err != nil {
		return err
	}

	if p.Folder == "" {
		return fmt.Errorf("empty foler")
	}

	nars, err := w.fs.List(p.Folder)
	if err != nil {
		return nil
	}

	for i, nar := range nars {
		b, err := w.tts.Synthesize(nar["text"].(string))
		if err != nil {
			return fmt.Errorf("tts failed idx = %s: %w", nar["id"].(string), err)
		}
		id := fmt.Sprintf("%04d", i)

		audioID, dst, err := w.fs.Save(p.Folder, id, ".wav", bytes.NewReader(b))
		if err != nil {
			return fmt.Errorf("save wav failed: %w", err)
		}
		fmt.Printf("id = %s, folder = %s\n", id, p.Folder)
		if err := w.fs.Add(p.Folder, id, map[string]any{"audio_id": id}, nil); err != nil {
			return fmt.Errorf("add field into %s's narrations failed: %w", p.Folder, err)
		}

		slog.Info("save wav ok",
			"folder", p.Folder,
			"autio_id", audioID,
			"dst", dst,
		)
	}

	return nil
}

func (w *worker) handleTTSSingle(task *Task) error {
	var p GenTTSSinglePayLoad

	if err := json.Unmarshal(task.Payload, &p); err != nil {
		return err
	}

	if p.Folder == "" {
		return fmt.Errorf("empty folder")
	}

	nars, err := w.fs.List(p.Folder)
	slog.Debug("fetch nars list from local store",
		"folder", p.Folder,
		"nars", nars,
	)
	if err != nil {
		return nil
	}

	for i, nar := range nars {
		if nar["id"].(string) != p.NarID {
			continue
		}

		b, err := w.tts.Synthesize(nar["text"].(string))
		if err != nil {
			return fmt.Errorf("tts failed idx = %s: %w", nar["id"].(string), err)
		}
		id := fmt.Sprintf("%04d", i)

		audioID, dst, err := w.fs.Save(p.Folder, id, ".wav", bytes.NewReader(b))
		if err != nil {
			return fmt.Errorf("save wav failed: %w", err)
		}
		fmt.Printf("id = %s, folder = %s\n", id, p.Folder)
		if err := w.fs.Add(p.Folder, id, map[string]any{"audio_id": id}, nil); err != nil {
			return fmt.Errorf("add field into %s's narrations failed: %w", p.Folder, err)
		}

		slog.Info("save wav ok",
			"folder", p.Folder,
			"autio_id", audioID,
			"dst", dst,
		)
	}

	return nil
}
