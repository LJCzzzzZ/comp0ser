package prompts_test

import (
	"fmt"
	"testing"

	"comp0ser/prompts"
)

func TestRender(t *testing.T) {
	r, err := prompts.NewRenderer()
	if err != nil {
		t.Fatal("failed to create renderer", err)
	}
	sys, err := r.System(prompts.Config{
		Subject:  "A",
		Segments: 1,
		MinChars: 1,
		MaxChars: 10,
		Focus:    "B",
		Hook:     "C",
	})
	if err != nil {
		t.Fatal("failed to gen system prompts", err)
	}

	fmt.Println(sys)
}
