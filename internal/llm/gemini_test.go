package llm

import (
	"context"
	"strings"
	"testing"

	"github.com/joho/godotenv"
)

func TestGemini_GenScript(t *testing.T) {
	godotenv.Load()
	client, err := NewClient(context.Background(), Config{APIKey: "AIzaSyCkaEB2Y5B7d9gbjbrvaO2XjiZLO4zA9KA"})
	if err != nil {
		t.Fatal(err)
	}
	model := "gemini-3-flash-preview"

	// 你要“多段文本”，这里让模型输出 3 段即可
	content := `请把下面内容拆成 3 段，保持原意但更口语：
今天我们要介绍一个新功能，它可以显著提升处理速度，同时保持稳定性。最后给出一句总结。`

	// 强约束：只输出 JSON 数组，每个元素是 string
	prompt := `你是一个只输出 JSON 的服务。
只输出合法 JSON，不要解释、不要 markdown、不要代码块。
输出必须是 JSON 数组（以 [ 开头，以 ] 结尾）。
数组每个元素都是字符串，表示一段文本。`

	nars, err := client.GenScript(context.Background(), model, content, prompt)
	if err != nil {
		t.Fatal(err)
	}
	if len(nars) == 0 {
		t.Fatalf("expected non-empty result")
	}
	for i, s := range nars {
		if strings.TrimSpace(s) == "" {
			t.Fatalf("segment %d is empty", i)
		}
	}
}
