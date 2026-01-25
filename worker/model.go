package worker

import (
	"context"
	"encoding/json"
)

type Cmd struct{}

type TaskType string

const (
	GenScript TaskType = "script.gen"
	GenTTSAll TaskType = "tts.all.gen"
)

type Task struct {
	ID      string
	Type    TaskType
	Payload json.RawMessage
}

type HandlerFunc func(ctx context.Context, payload json.RawMessage) (any, error)
