package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
)

func (w *worker) handleGensubtitle(task *Task) error {
	var p GenSubtitlePayload

	if err := json.Unmarshal(task.Payload, &p); err != nil {
		return err
	}

	slog.Info("gen subtitle task start")

	cmd, err := w.whisper.GenSubtitle(p.AudioPath, p.OutputPath, p.Lang)
	if err != nil {
		return fmt.Errorf("gen subtitle failed: %w", err)
	}

	fmt.Println(cmd.Args)

	if err := w.runner.Run(context.Background(), cmd); err != nil {
		return err
	}

	slog.Info("gen subtitle task ok",
		"audio_path", p.AudioPath,
		"output_path", p.OutputPath,
		"lang", p.Lang,
	)
	return nil
}

func (w *worker) handleBrunSubtitle(task *Task) error {
	var p BrunSubtitlePayLoad

	if err := json.Unmarshal(task.Payload, &p); err != nil {
		return err
	}

	slog.Info("brun subtitle task start")

	cmd, err := w.ff.BurnSubtitle(p.VideoPath, p.SubtitlePath, p.OutputPath)
	if err != nil {
		return fmt.Errorf("fetch cmd from brun subtitle failed: %w", err)
	}

	if err := w.runner.Run(context.Background(), cmd); err != nil {
		return err
	}

	slog.Info("gen subtitle task ok",
		"vedio_path", p.VideoPath,
		"subtitle_path", p.SubtitlePath,
		"output_paht", p.OutputPath,
	)

	return nil
}
