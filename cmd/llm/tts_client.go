package llm

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
)

const ttsURL = "https://openspeech.bytedance.com/api/v1/tts"

type TTSRequest struct {
	App     AppConfig   `json:"app"`
	User    UserConfig  `json:"user"`
	Audio   AudioConfig `json:"audio"`
	Request ReqConfig   `json:"request"`
}

type AppConfig struct {
	Cluster string `json:"cluster"`
	// 你现在 curl 没带 appid/token 也能用，这里就不强制
	// AppID string `json:"appid,omitempty"`
	// Token string `json:"token,omitempty"`
}

type UserConfig struct {
	UID string `json:"uid"`
}

type AudioConfig struct {
	VoiceType   string  `json:"voice_type"`
	Encoding    string  `json:"encoding"` // mp3 / wav / pcm...
	SpeedRatio  float64 `json:"speed_ratio"`
	Rate        int     `json:"rate,omitempty"`
	VolumeRatio float64 `json:"volume_ratio,omitempty"`
	PitchRatio  float64 `json:"pitch_ratio,omitempty"`
}

type ReqConfig struct {
	ReqID     string `json:"reqid"`
	Text      string `json:"text"`
	Operation string `json:"operation"` // query（HTTP只能 query） :contentReference[oaicite:3]{index=3}
	TextType  string `json:"text_type,omitempty"`
}

// -------- 返回结构 --------

type TTSResponse struct {
	ReqID     string          `json:"reqid"`
	Code      int             `json:"code"`
	Message   string          `json:"message"`
	Operation string          `json:"operation"`
	Sequence  int             `json:"sequence"`
	Data      string          `json:"data"` // base64 音频 :contentReference[oaicite:4]{index=4}
	Addition  json.RawMessage `json:"addition,omitempty"`
}

// -------- Client --------

type Client struct {
	APIKey     string       // 用你现在的 x-api-key
	HTTPClient *http.Client // 可注入自定义 client
}

func NewClient(apiKey string) *Client {
	return &Client{
		APIKey: apiKey,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Synthesize 返回 mp3 二进制内容
func (c *Client) Synthesize(ctx context.Context, cluster, uid, voiceType, text string) ([]byte, *TTSResponse, error) {
	if c.APIKey == "" {
		return nil, nil, errors.New("empty api key")
	}
	if cluster == "" || uid == "" || voiceType == "" || text == "" {
		return nil, nil, errors.New("cluster/uid/voiceType/text must not be empty")
	}

	reqBody := TTSRequest{
		App: AppConfig{
			Cluster: cluster,
		},
		User: UserConfig{
			UID: uid,
		},
		Audio: AudioConfig{
			VoiceType:  voiceType,
			Encoding:   "wav",
			SpeedRatio: 0.9,
			Rate:       24000,
		},
		Request: ReqConfig{
			ReqID:     uuid.NewString(), // 每次唯一 :contentReference[oaicite:5]{index=5}
			Text:      text,
			Operation: "query",
		},
	}

	bs, err := json.Marshal(reqBody)
	if err != nil {
		return nil, nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, ttsURL, bytes.NewReader(bs))
	if err != nil {
		return nil, nil, fmt.Errorf("new request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", c.APIKey)

	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("read body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, nil, fmt.Errorf("http status=%d body=%s", resp.StatusCode, string(raw))
	}

	var ttsResp TTSResponse
	if err := json.Unmarshal(raw, &ttsResp); err != nil {
		return nil, nil, fmt.Errorf("unmarshal response: %w, body=%s", err, string(raw))
	}

	// 3000 表示成功 :contentReference[oaicite:6]{index=6}
	if ttsResp.Code != 3000 {
		return nil, &ttsResp, fmt.Errorf("tts failed: code=%d message=%s", ttsResp.Code, ttsResp.Message)
	}

	// data 是 base64 音频，需要解码 :contentReference[oaicite:7]{index=7}
	audioBytes, err := base64.StdEncoding.DecodeString(ttsResp.Data)
	if err != nil {
		return nil, &ttsResp, fmt.Errorf("base64 decode audio: %w", err)
	}

	return audioBytes, &ttsResp, nil
}
