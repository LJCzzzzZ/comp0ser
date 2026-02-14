package cmd

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestWhsiper_GenSRT(t *testing.T) {
	w := NewWhisper(
		"/home/jincheng-l/code/subs/whisper.cpp/build/bin/whisper-cli",  /* bin path */
		"/home/jincheng-l/code/subs/whisper.cpp/models/ggml-medium.bin", /* model path */
	)

	cmd, err := w.GenSubtitle(
		"/mnt/media/data/neutron_star/audio/0000.wav", /* audio path */
		"/mnt/media/data/neutron_star/subtitle.srt",   /* srt output path */
		"zh",
	)

	fmt.Println(cmd)
	if err != nil {
		t.Fatal(err)
	}

	r := Runner{Timeout: 10 * time.Minute}
	if err := r.Run(context.Background(), cmd); err != nil {
		t.Fatal(err)
	}
	fmt.Println("gen srt finish")
}
