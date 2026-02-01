package worker

import (
	"context"
	"encoding/json"
)

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
	Brun         TaskType = "mp4.brun.sub"
)

type Task struct {
	ID      string
	Type    TaskType
	Payload json.RawMessage
}

type HandlerFunc func(ctx context.Context, payload json.RawMessage) (any, error)
