package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"comp0ser/config"
	"comp0ser/core"
	"comp0ser/core/provider"
	"comp0ser/filestore"
	"comp0ser/internal/api"
	"comp0ser/internal/logging"
	"comp0ser/prompts"
	"comp0ser/worker"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	if err := run(ctx); err != nil && !errors.Is(err, context.Canceled) {
		slog.Error("program exit with error",
			slog.Any("err", err),
		)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	_ = godotenv.Load()

	conf, err := config.Load()
	if err != nil {
		return fmt.Errorf("load confg failed: %w", err)
	}

	// init slog
	cleanupFn := logging.Init(logging.Config{
		Level:   conf.Level,
		Format:  conf.Format,
		LogFile: conf.LogFile,
		Silent:  conf.Silent,
		Debug:   conf.Debug,
	})

	defer cleanupFn()

	if conf.Debug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	tempRoot := "/tmp/comp0ser"
	if err := os.MkdirAll(tempRoot, 0o755); err != nil {
		return fmt.Errorf("mdkir temp root: %w", err)
	}
	defer func() {
		if err := os.RemoveAll(tempRoot); err != nil {
			slog.Error("remove temp root failed", "err", err, "tempRoot", tempRoot)
		} else {
			slog.Info("temp root cleaned", "tempRoot", tempRoot)
		}
	}()

	r, err := prompts.NewRenderer()
	if err != nil {
		return fmt.Errorf("init renderer: %w", err)
	}

	llmClient, err := provider.NewGeminiClient(ctx, conf.TextModel)
	if err != nil {
		return fmt.Errorf("init llm client: %w", err)
	}

	ttsClient := provider.NewTTSClient(provider.TTSConfig{
		APIKey:    conf.VolcAPIKey,
		Cluster:   conf.CLUSTER,
		UID:       conf.UID,
		VoiceType: conf.VoiceType,
		Timeout:   conf.Timeout,
	})

	fs := filestore.NewFileLocalStore(conf.LocalFileStoreDir)

	ff := core.NewFFmpeg("ffmpeg")

	runner := &core.Runner{Timeout: 30 * time.Minute}

	wk := worker.New(worker.Config{
		LLM:      llmClient,
		TTS:      ttsClient,
		Renderer: r,
		FS:       fs,
		FF:       ff,
		Runner:   runner,
	})
	wk.Start()
	defer wk.Shutdown()

	server := newHTTPServer(conf.Addr, wk)

	if err := serveHTTP(ctx, server); err != nil {
		return fmt.Errorf("serve http: %w", err)
	}
	return nil
}

func newHTTPServer(addr string, worker worker.Worker) *http.Server {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})

	api.NewCmdHandler(worker).Register(r)

	return &http.Server{
		Addr:    addr,
		Handler: r,
	}
}

func serveHTTP(ctx context.Context, srv *http.Server) error {
	errCh := make(chan error, 1)
	go func() {
		slog.Info("http server started", "addr", srv.Addr)
		errCh <- srv.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		if err == nil || errors.Is(err, http.ErrServerClosed) {
			slog.Info("http server stopped")
			return nil
		}
		return err
	case <-ctx.Done():
		slog.Info("shutdown signal received")

		sdCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := srv.Shutdown(sdCtx); err != nil {
			return fmt.Errorf("http shutdown: %w", err)
		}

		err := <-errCh
		if err == nil || errors.Is(err, http.ErrServerClosed) {
			slog.Info("shutdown complete")
			return nil
		}
		return err
	}
}
