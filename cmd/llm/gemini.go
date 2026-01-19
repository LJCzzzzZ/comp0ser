// Package llm provide llm client for composer
package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"os"
	"strings"

	"google.golang.org/genai"
)

type GeminiClient struct {
	client *genai.Client
}

func NewGeminiClient(ctx context.Context) (*GeminiClient, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY is empty")
	}

	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: apiKey,
	})
	if err != nil {
		return nil, fmt.Errorf("genai.NewClient failed: %w", err)
	}

	return &GeminiClient{client: client}, nil
}

func (g *GeminiClient) GenScript(ctx context.Context, model, content, prompt string) ([]string, error) {
	if model == "" {
		return nil, fmt.Errorf("model is empty")
	}

	if strings.TrimSpace(content) == "" {
		return nil, fmt.Errorf("content is empty")
	}

	config := &genai.GenerateContentConfig{
		SystemInstruction: &genai.Content{
			Parts: []*genai.Part{
				{
					Text: ScriptSystemPrompt,
				},
			},
		},
	}

	data := string(content)
	slog.Info("request genmini llm to gen script", slog.Int("data_len", len(data)))
	resp, err := g.client.Models.GenerateContent(
		ctx,
		model,
		genai.Text(data),
		config,
	)
	if err != nil {
		log.Fatal(err)
	}

	ret, err := parseSegments(resp.Text())
	if err != nil {
		return nil, err
	}
	return ret, nil
}

type ScriptSegments struct {
	Segments []string `json:"segments"`
}

func parseSegments(narration string) ([]string, error) {
	// 去掉可能的 ```json ``` 包裹
	narration = strings.TrimSpace(narration)
	narration = strings.TrimPrefix(narration, "```json")
	narration = strings.TrimPrefix(narration, "```")
	narration = strings.TrimSuffix(narration, "```")
	narration = strings.TrimSpace(narration)

	var s ScriptSegments
	if err := json.Unmarshal([]byte(narration), &s); err != nil {
		return nil, fmt.Errorf("unmarshal segments json failed: %w\nraw=%s", err, narration)
	}
	if len(s.Segments) == 0 {
		return nil, fmt.Errorf("segments is empty")
	}
	return s.Segments, nil
}
