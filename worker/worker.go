package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"comp0ser/core"
	"comp0ser/core/provider"
	"comp0ser/filestore"
	"comp0ser/prompts"
)

const (
	defaultWorkerCount   = 5
	defaultQueueCapacity = defaultWorkerCount * 8
)

type Config struct {
	WorkerCount   int
	QueueCapacity int

	FF     *core.FFmpeg
	Runner *core.Runner

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
	ff       *core.FFmpeg
	llm      *provider.GeminiClient
	tts      *provider.TTSClient
	renderer *prompts.Renderer
	runner   *core.Runner

	workerCount   int
	queueCapacity int

	wg        sync.WaitGroup
	startOnce sync.Once
	stopOnce  sync.Once

	mu     sync.RWMutex
	closed bool

	queue chan *Task
}

func New(conf Config) Worker {
	wc := conf.WorkerCount
	if wc <= 0 {
		wc = defaultWorkerCount
	}
	qc := conf.QueueCapacity
	if qc <= 0 {
		qc = defaultQueueCapacity
	}
	return &worker{
		fs:            conf.FS,
		runner:        conf.Runner,
		ff:            conf.FF,
		llm:           conf.LLM,
		tts:           conf.TTS,
		renderer:      conf.Renderer,
		workerCount:   wc,
		queueCapacity: qc,
	}
}

func (w *worker) Start() {
	w.startOnce.Do(func() {
		slog.Info(
			"worker starting",
			"workers", w.workerCount,
			"queueCapacity", w.queueCapacity,
		)
		w.queue = make(chan *Task, w.queueCapacity)

		for i := 0; i < w.workerCount; i++ {
			w.wg.Go(func() {
				w.loop()
			})
		}
	})
}

func (w *worker) Shutdown() {
	w.stopOnce.Do(func() {
		slog.Info("worker shutting down")

		w.mu.Lock()
		w.closed = true
		if w.queue != nil {
			close(w.queue)
		}
		w.mu.Unlock()

		w.wg.Wait()
		slog.Info("worker stopped")
	})
}

func (w *worker) Submit(ctx context.Context, typ TaskType, payload any) (string, error) {
	w.Start()

	b, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	taskID := fmt.Sprintf("task_%d", time.Now().UnixNano())

	task := &Task{
		ID:      taskID,
		Type:    typ,
		Payload: b,
	}
	w.mu.RLock()
	defer w.mu.RUnlock()

	select {
	case w.queue <- task:
		return taskID, nil
	case <-ctx.Done():
		return "", ctx.Err()
	}
}

func (w *worker) loop() {
	for task := range w.queue {
		if err := w.runOneSafe(task); err != nil {
			slog.Error("run task failed",
				"task_id", task.ID,
				"task_type", task.Type,
				"err", err,
			)
		}
	}
}

func (w *worker) runOneSafe(task *Task) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v", r)
			slog.Error("task panicked",
				"task_id", task.ID,
				"task_type", task.Type,
				"err", err,
			)
		}
	}()
	return w.runOne(task)
}

func (w *worker) runOne(task *Task) error {
	switch task.Type {
	case GenScript:
		return w.handleScriptGen(task)
	case GenTTSAll:
		return w.handleTTSAll(task)
	case GenTTSSingle:
		return w.handleTTSSingle(task)
	case Mixdown:
		return w.handleMixdown(task)
	case Concat:
		return w.handleConcat(task)
	case Render:
		return w.handleRender(task)
	case Merge:
		return w.handleMerge(task)
	case Brun:
		return w.handleBrunSubtitle(task)
	default:
		return fmt.Errorf("unknown task type: %v", task.Type)
	}
}
