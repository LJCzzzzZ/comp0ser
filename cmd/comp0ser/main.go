package main

import (
	"flag"
	"log/slog"
	"os"

	"comp0ser/internal/app"
	"comp0ser/internal/logging"
)

var (
	logLevel, logMode               string
	LLMAPIKey, TTSAPIKey, voiceType string

	storeDir, tmpRoot        string
	whisperBin, whisperModel string

	port string
)

func main() {
	flag.StringVar(&logLevel, "log_level", envOr("LOG_LEVLE", "info"), "log level: debug|info")
	flag.StringVar(&logMode, "log_mode", envOr("LOG_MODE", "dev"), "log mode: dev|prd")
	flag.StringVar(&LLMAPIKey, "llm_api_key", envOr("GEMINI_API_KEY", ""), "gemini api key")
	flag.StringVar(&TTSAPIKey, "tts_api_key", envOr("VOLC_API_KEY", ""), "volc api key")
	flag.StringVar(&storeDir, "store_dir", envOr("STORE_DIR", "/store"), "file local store dir")
	flag.StringVar(&port, "port", envOr("SERVER_PORT", "8088"), "http server port")
	flag.StringVar(&voiceType, "voice_type", envOr("VOICE_TYPE", ""), "volc tts voice type field")
	flag.StringVar(&tmpRoot, "tmp_root", envOr("TMP_ROOT", "/tmp/comp0ser"), "temp file root")
	flag.StringVar(&whisperBin, "whisper_bin", envOr("WHISPER_BIN", ""), "whisper bin path")
	flag.StringVar(&whisperModel, "whisper_model", envOr("WHISPER_MODEL", ""), "whisper model path")
	flag.Parse()

	logger := logging.NewLogger(logLevel, logMode)
	slog.SetDefault(logger)

	slog.Info("app start")

	if err := app.Run(app.Options{
		LLMAPIKey:    LLMAPIKey,
		TTSAPIKey:    TTSAPIKey,
		StoreDir:     storeDir,
		Port:         port,
		VoiceType:    voiceType,
		TmpRoot:      tmpRoot,
		WhisperBin:   whisperBin,
		WhisperModel: whisperModel,
	}); err != nil {
		slog.Error("application exit",
			"err", err,
		)
	}
}

func envOr(key, def string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return def
}
