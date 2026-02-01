// Package config provide config
package config

import (
	"fmt"
	"time"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	Addr string `env:"ADDR" envDefault:":8088"`

	// Log Config
	LogFile string `env:"LOG_FILE" envDefault:"run.log"`
	Level   string `env:"LEVEL" envDefault:"info"`
	Format  string `env:"FORMAT" envDefault:"json"`
	Silent  bool   `env:"SILENT" envDefault:"false"`
	Debug   bool   `env:"DEBUG" envDefault:"false"`

	// Gemini Config
	GeminiAPIKey string `env:"GEMINI_API_KEY" envRequired:"true"`
	TextModel    string `env:"TEXT_MODEL" envDefault:"gemini-3-flash-preview"`

	// Volc Config
	VolcAPIKey string        `env:"VOLC_API_KEY" envRequired:"true"`
	CLUSTER    string        `env:"CLUSTER" envDefault:"volcano_icl"`
	UID        string        `env:"UID" envDefault:"comp0ser"`
	VoiceType  string        `env:"VOICE_TYPE" envRequired:"true"`
	Timeout    time.Duration `env:"TIMEOUT" envDefault:"30m"`

	// Store Config
	LocalFileStoreDir string `env:"LOCAL_FILE_STORE_DIR" envDefault:"/mnt/media/data"`

	// Audio Config
	Volume float64 `env:"VOLUME" envDefault:"0.18"`
}

func Load() (*Config, error) {
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return nil, err
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (c *Config) Validate() error {
	if c.Addr == "" {
		return fmt.Errorf("SERVER_ADDR is required")
	}

	if c.Level == "" {
		return fmt.Errorf("LOG LEVEL is required")
	}

	if c.TextModel == "" {
		return fmt.Errorf("TEXT MODEL is required")
	}
	return nil
}
