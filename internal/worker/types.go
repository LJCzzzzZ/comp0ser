package worker

import (
	"context"
	"encoding/json"
)

type GenScriptPayLoad struct {
	RawText  string `json:"rawText"`
	Subject  string `json:"subject"`
	Segments int    `json:"segments"`
	MinChars int    `json:"minChars"`
	MaxChars int    `json:"maxChars"`

	// focus for generated narrations
	Focus string `json:"focus"`

	Hook string `json:"hook"`

	Model string `json:"model"`
}

type GenSubtitlePayload struct {
	AudioPath  string `json:"audioPath"`
	OutputPath string `json:"outputPath"`
	Lang       string `json:"lang"`
}

type RenderPayLoad struct {
	Folder  string  `jsonm:"foler"`
	Dur     float64 `json:"dur"`     // 目标总时长（秒）
	TailCut float64 `json:"tailCut"` // 每段末尾剪掉秒数（比如 10）
	Loop    bool    `json:"loop"`    // 不够 dur 是否循环补足
	Out     string  `json:"out"`     // 可选：输出文件名
}

type BrunSubtitlePayLoad struct {
	VideoPath    string `json:"videoPath"`
	SubtitlePath string `json:"subtitlePath"`
	OutputPath   string `json:"outputPath"`
}

type MergePayLoad struct {
	VideoPath string `json:"videoPath"`
	AudioPath string `json:"audioPath"`
	OutPath   string `json:"outPath"`
}

type ConcatPayLoad struct {
	Folder string `jsonm:"foler"`
}

type GenTTSPayLoad struct {
	Folder string `json:"folder"`
}

type GenTTSSinglePayLoad struct {
	Folder string `json:"folder"`
	NarID  string `json:"narId"`
}

type MixdownPayLoad struct {
	AudioPath string  `json:"audioPath"`
	BGMPath   string  `json:"BGMPath"`
	Filename  string  `json:"filename"`
	Volume    float64 `json:"volume"`
	Loop      bool    `json:"loop"`
}

type Narration struct {
	ID      int    `json:"id"`
	Text    string `json:"text"`
	AudioID string `json:"audio_id"`
}
type Cmd struct{}

type TaskType string

const (
	GenScript    TaskType = "script.gen"
	GenTTSAll    TaskType = "tts.all.gen"
	GenTTSSingle TaskType = "tts.single.gen"
	Mixdown      TaskType = "mix.audio.bgm"
	Concat       TaskType = "concat.wav"
	Render       TaskType = "render.mp4"
	Merge        TaskType = "m4a.merge.mp4"
	GenSrt       TaskType = "audio.gen.srt"
	Brun         TaskType = "mp4.brun.sub"
)

type Task struct {
	ID      string
	Type    TaskType
	Payload json.RawMessage
}

type HandlerFunc func(ctx context.Context, payload json.RawMessage) (any, error)
