package cmd

import (
	"context"
	"testing"
	"time"
)

func TestFFmpeg_Probe_Duration(t *testing.T) {
	ff := NewFFmpeg("ffmpeg")

	cmd, err := ff.ConcatAssets([]string{
		"/mnt/media/data/a.mp4",
		"/mnt/media/data/b.mp4",
	}, "out.mp4", 600, 10, true)
	if err != nil {
		t.Fatal(err)
	}
	r := &Runner{Timeout: 10 * time.Minute}
	if err := r.Run(context.Background(), cmd); err != nil {
		t.Fatal(err)
	}
}
