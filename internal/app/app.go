package app

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"
	"time"

	"comp0ser/internal/cmd"
	"comp0ser/internal/filestore"
	"comp0ser/internal/llm"
	"comp0ser/internal/server"
	"comp0ser/internal/tts"
	"comp0ser/internal/worker"
	"comp0ser/prompts"
)

type Options struct {
	LLMAPIKey string
	TTSAPIKey string
	VoiceType string

	Model string

	StoreDir string
	TmpRoot  string

	WhisperBin   string
	WhisperModel string

	Port string
}

func Run(opts Options) error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	defer func() {
		if r := recover(); r != nil {
			slog.Error("application panic",
				"panic", r,
				"stack", string(debug.Stack()),
			)
			os.Exit(1)
		}
	}()

	return run(ctx, opts)
}

func run(ctx context.Context, opts Options) error {
	llmClient, err := llm.NewClient(ctx, llm.Config{APIKey: opts.LLMAPIKey})
	if err != nil {
		return fmt.Errorf("create gemini client: %w", err)
	}

	ttsClient, err := tts.NewClient(
		tts.WithAPIKey(opts.TTSAPIKey),
		tts.WithVoiceType(opts.VoiceType),
	)
	if err != nil {
		return fmt.Errorf("create tts client: %w", err)
	}

	fs := filestore.NewFileLocalStore(opts.StoreDir)
	slog.Info("file local store init",
		"dir", opts.StoreDir,
	)

	ff := cmd.NewFFmpeg("ffmpeg")
	whisper := cmd.NewWhisper(opts.WhisperBin, opts.WhisperModel)

	renderer, err := prompts.NewRenderer()
	if err != nil {
		return fmt.Errorf("create renderer: %w", err)
	}

	runner := &cmd.Runner{
		Timeout: 60 * time.Minute,
	}

	wk := worker.New(worker.Config{
		FS:       fs,
		FF:       ff,
		LLM:      llmClient,
		TTS:      ttsClient,
		Renderer: renderer,
		Runner:   runner,
		Whisper:  whisper,
	})

	if err := os.MkdirAll(opts.TmpRoot, 0o755); err != nil {
		return fmt.Errorf("init tmp root failed: %w", err)
	}

	defer func() {
		_ = os.RemoveAll(opts.TmpRoot)
	}()

	srv, err := server.New(opts.Port)
	if err != nil {
		return fmt.Errorf("server.New: %w", err)
	}
	slog.Info("server listening", "port", opts.Port)

	srv.ServerHTTPHandler(ctx, server.Routes(ctx, wk, opts.TmpRoot))
	return nil
}
