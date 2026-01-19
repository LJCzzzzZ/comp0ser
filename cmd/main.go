package main

import (
	"context"
	"encoding/binary"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"path/filepath"

	"comp0ser/config"

	"github.com/joho/godotenv"
)

func main() {
	if err := run(); err != nil {
		slog.Error("program exit with error",
			slog.Any("err", err),
			slog.String("err_chain", errChain(err)),
		)
		os.Exit(1)
	}
}

func run() error {
	conf := loadConf()

	cleanup := initLogger(conf)
	defer cleanup()

	if err := loadEnv(); err != nil {
		return err
	}

	composer, err := NewComposer(conf)
	if err != nil {
		return err
	}
	if err := composer.Work(context.Background()); err != nil {
		return err
	}
	return nil
}

func loadConf() *config.Config {
	var conf config.Config
	if err := conf.Load(); err != nil {
		if err == flag.ErrHelp {
			os.Exit(0)
		}
		log.Fatalln("load config failed ", err)
	}
	return &conf
}

func initLogger(conf *config.Config) func() {
	level := slog.LevelInfo
	if conf.Debug {
		level = slog.LevelDebug
	}

	var w io.Writer = os.Stdout
	if conf.Silent {
		w = io.Discard
	}

	var file *os.File
	if conf.LogFile != "" {
		_ = os.MkdirAll(filepath.Dir(conf.LogFile), 0o755)

		f, err := os.OpenFile(conf.LogFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
		if err == nil {
			file = f

			if conf.Silent {
				w = file
			} else {
				w = io.MultiWriter(os.Stdout, file)
			}
		} else {
			fmt.Fprintf(os.Stderr, "open log file failed: %v (fallback to stdout)\n", err)
		}
	}

	handler := slog.NewTextHandler(w, &slog.HandlerOptions{
		Level:     level,
		AddSource: conf.Debug,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				t := a.Value.Time()
				a.Value = slog.StringValue(t.Format("2006-01-02 15:04:05"))
			}
			return a
		},
	})

	slog.SetDefault(slog.New(handler))

	return func() {
		if file != nil {
			_ = file.Close()
		}
	}
}

func writeWav(filename string, pcm []byte, channels uint16, sampleRate uint32, bitsPerSample uint16) error {
	// WAV/RIFF header sizes
	blockAlign := channels * (bitsPerSample / 8)
	byteRate := sampleRate * uint32(blockAlign)
	dataSize := uint32(len(pcm))
	riffSize := 36 + dataSize // 4 + (8 + SubChunk1) + (8 + SubChunk2) = 36 + data

	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	// RIFF header
	if _, err := f.Write([]byte("RIFF")); err != nil {
		return err
	}
	if err := binary.Write(f, binary.LittleEndian, riffSize); err != nil {
		return err
	}
	if _, err := f.Write([]byte("WAVE")); err != nil {
		return err
	}

	// fmt chunk
	if _, err := f.Write([]byte("fmt ")); err != nil {
		return err
	}
	if err := binary.Write(f, binary.LittleEndian, uint32(16)); err != nil { // PCM fmt chunk size
		return err
	}
	if err := binary.Write(f, binary.LittleEndian, uint16(1)); err != nil { // AudioFormat 1 = PCM
		return err
	}
	if err := binary.Write(f, binary.LittleEndian, channels); err != nil {
		return err
	}
	if err := binary.Write(f, binary.LittleEndian, sampleRate); err != nil {
		return err
	}
	if err := binary.Write(f, binary.LittleEndian, byteRate); err != nil {
		return err
	}
	if err := binary.Write(f, binary.LittleEndian, blockAlign); err != nil {
		return err
	}
	if err := binary.Write(f, binary.LittleEndian, bitsPerSample); err != nil {
		return err
	}

	// data chunk
	if _, err := f.Write([]byte("data")); err != nil {
		return err
	}
	if err := binary.Write(f, binary.LittleEndian, dataSize); err != nil {
		return err
	}
	if _, err := f.Write(pcm); err != nil {
		return err
	}

	return nil
}

func loadEnv() error {
	godotenv.Load()
	return nil
}

func errChain(err error) string {
	out := ""
	for i := 0; err != nil && i < 12; i++ {
		out += fmt.Sprintf("[%d] %T: %v\n", i, err, err)
		err = errors.Unwrap(err)
	}
	return out
}
