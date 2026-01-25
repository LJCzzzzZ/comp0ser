package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"comp0ser/core/provider"
	"comp0ser/filestore"
	"comp0ser/prompts"
)

const maxTasks = 5

type Config struct {
	// file store

	FS       filestore.FileStore
	LLM      *provider.GeminiClient
	TTS      *provider.TTSClient
	Renderer *prompts.Renderer
}

// Worker defines interface for excutor
type Worker interface {
	Start()
	Shutdown()
	Submit(context.Context, TaskType, any) (string, error)
}

type worker struct {
	fs       filestore.FileStore
	llm      *provider.GeminiClient
	tts      *provider.TTSClient
	renderer *prompts.Renderer

	wg        sync.WaitGroup
	startOnce sync.Once
	stopOnce  sync.Once

	queue chan *Task
	done  chan struct{}
}

func New(conf Config) Worker {
	return &worker{
		fs:       conf.FS,
		llm:      conf.LLM,
		tts:      conf.TTS,
		renderer: conf.Renderer,
	}
}

func (w *worker) Start() {
	w.startOnce.Do(func() {
		slog.Info(
			"worker starting",
			"maxTasks", maxTasks,
		)
		w.queue = make(chan *Task, maxTasks)
		w.done = make(chan struct{})

		w.wg.Go(func() { w.loop() })
	})
}

func (w *worker) Shutdown() {
	w.stopOnce.Do(func() {
		slog.Info("worker shutting down")
		close(w.done)
		w.wg.Wait()
		slog.Info("worker stopped")
	})
}

func (w *worker) Submit(ctx context.Context, typ TaskType, playload any) (string, error) {
	b, err := json.Marshal(playload)
	if err != nil {
		return "", err
	}
	taskID := fmt.Sprintf("task_%d", time.Now().UnixNano())

	task := &Task{
		ID:      taskID,
		Type:    typ,
		Payload: b,
	}

	select {
	case w.queue <- task:
		return taskID, nil
	case <-ctx.Done():
		return "", ctx.Err()
	}
}

func (w *worker) loop() {
	for {
		select {
		case task, ok := <-w.queue:
			if !ok {
				return
			}

			if err := w.runOne(task); err != nil {
				slog.Error("run task failed",
					"task_id", task.ID,
					"task_type", task.Type,
					"err", err,
				)
				continue
			}

		case <-w.done:
			return
		}
	}
}

func (w *worker) runOne(task *Task) error {
	switch task.Type {
	case GenScript:
		return w.handleScriptGen(task)
	case GenTTSAll:
		return w.handleTTSAll(task)
	default:
		return fmt.Errorf("unknown task type: %v", task.Type)
	}
}
