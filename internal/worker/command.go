package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

func (w *worker) handleMerge(task *Task) error {
	var p MergePayLoad

	if err := json.Unmarshal(task.Payload, &p); err != nil {
		return err
	}

	slog.Info("merge task start")

	cmd, err := w.ff.Merge(p.VideoPath, p.AudioPath, p.OutPath)
	if err != nil {
		return err
	}

	if err := w.runner.Run(context.Background(), cmd); err != nil {
		return err
	}

	slog.Info("merge task finish")
	return nil
}

func (w *worker) handleMixdown(task *Task) error {
	log := slog.With(
		"worker_handler", "mixdown",
		"taskID", task.ID,
	)
	var p MixdownPayLoad
	if err := json.Unmarshal(task.Payload, &p); err != nil {
		return err
	}

	log.Info("mixdown task start")

	cmd := w.ff.BlendM4A(p.AudioPath, p.BGMPath, p.Filename, p.Volume, p.Loop)
	fmt.Println(cmd)

	ctx, cannel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cannel()

	if err := w.runner.Run(ctx, cmd); err != nil {
		log.Error("runner execution failed",
			"err", err,
			"cmd", cmd,
		)
		return err
	}

	slog.Info("blend M4A, finished")
	return nil
}

func (w *worker) handleConcat(task *Task) error {
	var p ConcatPayLoad

	if err := json.Unmarshal(task.Payload, &p); err != nil {
		return err
	}

	slog.Info("concat task start")
	fmt.Println(p.Folder)

	nars, err := w.fs.List(p.Folder)
	if err != nil {
		return err
	}
	wavs := make([]string, len(nars))
	for i, nar := range nars {
		wavs[i] = filepath.Join(w.fs.Dir(), p.Folder, "audio", nar["audio_id"].(string)+".wav")
	}

	cmd, err := w.ff.ConcatWav(wavs, filepath.Join(w.fs.Dir(), p.Folder), p.Folder+".wav")
	if err != nil {
		return err
	}

	if err := w.runner.Run(context.Background(), cmd); err != nil {
		return err
	}

	slog.Info("concat wav finish",
		"len", len(wavs),
	)
	return nil
}

func (w *worker) handleRender(task *Task) error {
	var p RenderPayLoad

	if err := json.Unmarshal(task.Payload, &p); err != nil {
		return err
	}

	fmt.Printf("%+v\n", p)

	slog.Info("render task start", "folder", p.Folder)

	assetDir := filepath.Join(w.fs.Dir(), p.Folder, "asset")
	videos, err := listMP4Files(assetDir)
	fmt.Println(videos)
	if err != nil {
		return fmt.Errorf("list mp4 failed: %w", err)
	}
	if len(videos) == 0 {
		return fmt.Errorf("no mp4 files found in %s", assetDir)
	}

	out := p.Out
	if out == "" {
		out = "out.mp4"
	}
	outPath := filepath.Join(w.fs.Dir(), p.Folder, out)

	tailCut := p.TailCut
	if tailCut <= 0 {
		tailCut = 10
	}

	cmd, err := w.ff.ConcatAssets(videos, outPath, p.Dur, tailCut, p.Loop)

	fmt.Println(cmd)
	if err != nil {
		return err
	}

	if err := w.runner.Run(context.Background(), cmd); err != nil {
		return err
	}

	slog.Info("render task finish")
	return nil
}

func listMP4Files(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var files []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if strings.EqualFold(filepath.Ext(name), ".mp4") {
			files = append(files, filepath.Join(dir, name))
		}
	}

	// 保证拼接顺序稳定：按文件名排序
	sort.Strings(files)
	return files, nil
}
