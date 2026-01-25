package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"comp0ser/config"
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
	if err := execute(os.Args[1:]); err != nil {
		slog.Error("program exit with error",
			slog.Any("err", err),
		)
		os.Exit(1)
	}
}

func execute(args []string) error {
	loadEnv()

	conf, err := config.Load(args)
	if err != nil {
		if errors.Is(err, config.ErrHelp) {
			return nil
		}
		return err
	}

	// init slog
	cleanupFn := logging.Init(logging.Config{
		Level:   conf.LogLevel,
		Format:  conf.LogFormat,
		LogFile: conf.LogFile,
		Silent:  conf.Silent,
		Debug:   conf.Debug,
	})

	defer cleanupFn()

	r, err := prompts.NewRenderer()
	if err != nil {
		return err
	}

	llmClient, err := provider.NewGeminiClient(context.Background(), "gemini-3-flash-preview")
	if err != nil {
		return err
	}

	ttsClient := provider.NewTTSClient(provider.TTSConfig{
		APIKey:    os.Getenv("VOLC_API_KEY"),
		Cluster:   "volcano_icl",
		UID:       "comp0ser",
		VoiceType: "S_L7R26kdR1",
	})

	fs := filestore.NewFileLocalStore("/mnt/media/data")

	worker := newWorker(conf, r, llmClient, ttsClient, fs)
	worker.Start()
	defer worker.Shutdown()

	server := newHTTPServer(conf.HTTPAddr, worker)
	go func() {
		slog.Info("http server started", "addr", conf.HTTPAddr)
		if err := server.ListenAndServe(); err != nil {
			slog.Error("http server error", "err", err)
		}
	}()

	// exit graceful
	waitSignal()

	return nil
}

func loadEnv() {
	godotenv.Load()
}

func newWorker(conf *config.Config, r *prompts.Renderer, llm *provider.GeminiClient, tts *provider.TTSClient, fs filestore.FileStore) worker.Worker {
	w := worker.New(worker.Config{
		LLM:      llm,
		TTS:      tts,
		Renderer: r,
		FS:       fs,
	})
	return w
}

func newHTTPServer(addr string, worker worker.Worker) *http.Server {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})

	cmdHandle := api.NewCmdHandler(worker)
	cmdHandle.Register(r)
	return &http.Server{
		Addr:    addr,
		Handler: r,
	}
}

func waitSignal() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
}
