package tts

import (
	"os"
	"testing"
)

func TestVolcSythesize(t *testing.T) {
	client, err := NewClient(WithAPIKey("8a6f5c77-20da-4228-a409-a01f467829c2"), WithVoiceType("S_L7R26kdR1"))
	if err != nil {
		t.Fatal(err)
	}
	data, err := client.Synthesize("2007年，邓肯·洛里默等人在澳大利亚帕克斯电波天文台2001年的档案资料里发现了洛里默爆发")
	if err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile("output.wav", data, 0o644); err != nil {
		t.Fatal(err)
	}
}
