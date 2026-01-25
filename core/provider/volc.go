package provider

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

const ttsURL = "https://openspeech.bytedance.com/api/v1/tts"

type TTSConfig struct {
	APIKey string

	Cluster   string
	UID       string
	VoiceType string

	Timeout time.Duration
}

type TTSClient struct {
	apiKey string

	cluster   string
	uid       string
	voiceType string

	HTTPClient *http.Client
}

func NewTTSClient(conf TTSConfig) *TTSClient {
	return &TTSClient{
		apiKey:    conf.APIKey,
		cluster:   conf.Cluster,
		uid:       conf.Cluster,
		voiceType: conf.VoiceType,

		HTTPClient: &http.Client{
			Timeout: conf.Timeout,
		},
	}
}

func (c *TTSClient) Synthesize(ctx context.Context, text string) ([]byte, error) {
	if strings.TrimSpace(text) == "" {
		return nil, fmt.Errorf("text is empty")
	}

	reqBody := TTSRequest{
		App: AppConfig{
			Cluster: c.cluster,
		},
		User: UserConfig{
			UID: c.uid,
		},
		Audio: AudioConfig{
			VoiceType:  c.voiceType,
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
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, ttsURL, bytes.NewReader(bs))
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", c.apiKey)

	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http status=%d body=%s", resp.StatusCode, string(raw))
	}

	var ttsResp TTSResponse
	if err := json.Unmarshal(raw, &ttsResp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w, body=%s", err, string(raw))
	}

	if ttsResp.Code != 3000 {
		return nil, fmt.Errorf("tts failed: code=%d message=%s", ttsResp.Code, ttsResp.Message)
	}

	audioBytes, err := base64.StdEncoding.DecodeString(ttsResp.Data)
	if err != nil {
		return nil, fmt.Errorf("base64 decode audio: %w", err)
	}

	return audioBytes, nil
}
