package tts

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
)

const (
	FormatWAV = "wav"
	FormatPCM = "pcm"
	FormatMP3 = "mp3"

	SampleRate24K = 24000
)

const (
	_defailtEndpoint = "https://openspeech.bytedance.com/api/v1/tts"
	_defaultCluster  = "volcano_icl"
	_defaultUID      = "comp0ser"
	_defaultFormat   = FormatWAV
)

type options struct {
	Endpoint string
	APIKey   string

	AppID     string
	Cluster   string
	UID       string
	VoiceType string
	Format    string
}

type Option func(opts *options)

type Client interface {
	Synthesize(content string) ([]byte, error)
}

type client struct {
	opts options

	// Internal http client
	cli *http.Client
}

func defaultOpts() options {
	return options{
		Endpoint: _defailtEndpoint,
		Format:   _defaultFormat,
		UID:      _defaultUID,
		Cluster:  _defaultCluster,
	}
}

func NewClient(opts ...Option) (Client, error) {
	o := defaultOpts()

	for _, opt := range opts {
		opt(&o)
	}
	c := &http.Client{
		Timeout: 30000 * time.Millisecond,
	}
	return &client{opts: o, cli: c}, nil
}

func (c *client) Synthesize(content string) ([]byte, error) {
	reqID := uuid.NewString()
	var rb SynthesizeReq

	rb.User.UID = c.opts.UID
	rb.App.Cluster = c.opts.Cluster
	{
		rb.Audio.VoiceType = c.opts.VoiceType
		rb.Audio.Encoding = c.opts.Format
		rb.Audio.SpeedRatio = 1.0
	}
	{
		rb.Request.ReqID = reqID
		rb.Request.Text = content
		rb.Request.Operation = "query"
	}

	fmt.Println(rb)
	fmt.Println("Synthesize")

	body, err := json.Marshal(&rb)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, c.opts.Endpoint, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.opts.APIKey)

	resp, err := c.cli.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode/100 != 2 {
		return nil, fmt.Errorf("http %d", resp.StatusCode)
	}

	var r SynthesizeResp
	if err := json.Unmarshal(raw, &r); err != nil {
		return nil, err
	}

	if r.Code != 3000 {
		if r.Message != "" {
			return nil, fmt.Errorf("resp code=%d msg=%s", r.Code, r.Message)
		}
		return nil, fmt.Errorf("resp code=%d", r.Code)
	}

	audio, err := base64.StdEncoding.DecodeString(r.Data)
	if err != nil {
		return nil, err
	}
	return audio, nil
}

func WithAPIKey(v string) Option {
	return func(opts *options) { opts.APIKey = v }
}

func WithAppID(v string) Option {
	return func(opts *options) { opts.AppID = v }
}

func WithCluster(v string) Option {
	return func(opts *options) { opts.Cluster = v }
}

func WithUID(v string) Option {
	return func(opts *options) { opts.UID = v }
}

func WithVoiceType(v string) Option {
	return func(opts *options) { opts.VoiceType = v }
}
