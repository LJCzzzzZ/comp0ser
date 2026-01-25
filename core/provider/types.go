package provider

import "encoding/json"

const (
	DefaultTTSURL     = "https://openspeech.bytedance.com/api/v1/tts"
	DefaultEncoding   = "wav"
	DefaultSpeedRatio = 1.0
)

type TTSRequest struct {
	App     AppConfig   `json:"app"`
	User    UserConfig  `json:"user"`
	Audio   AudioConfig `json:"audio"`
	Request ReqConfig   `json:"request"`
}

type AppConfig struct {
	Cluster string `json:"cluster"`
}

type UserConfig struct {
	UID string `json:"uid"`
}

type AudioConfig struct {
	VoiceType   string  `json:"voice_type"`
	Encoding    string  `json:"encoding"` // mp3 / wav / pcm ...
	SpeedRatio  float64 `json:"speed_ratio"`
	Rate        int     `json:"rate,omitempty"`
	VolumeRatio float64 `json:"volume_ratio,omitempty"`
	PitchRatio  float64 `json:"pitch_ratio,omitempty"`
}

type ReqConfig struct {
	ReqID     string `json:"reqid"`
	Text      string `json:"text"`
	Operation string `json:"operation"`
	TextType  string `json:"text_type,omitempty"`
}

// -------- 返回结构 --------

type TTSResponse struct {
	ReqID     string          `json:"reqid"`
	Code      int             `json:"code"`
	Message   string          `json:"message"`
	Operation string          `json:"operation"`
	Sequence  int             `json:"sequence"`
	Data      string          `json:"data"` // base64 音频
	Addition  json.RawMessage `json:"addition,omitempty"`
}
