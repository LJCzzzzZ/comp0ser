package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"google.golang.org/genai"
)

type Config struct {
	APIKey string
}

type GeminiClient struct {
	client *genai.Client
}

func NewClient(ctx context.Context, conf Config) (*GeminiClient, error) {
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: conf.APIKey,
	})
	if err != nil {
		return nil, err
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

	slog.Info("request genmini llm",
		"model", model,
		"data_len", len(content),
	)
	resp, err := g.client.Models.GenerateContent(
		ctx,
		model,
		genai.Text(content),
		&genai.GenerateContentConfig{
			SystemInstruction: &genai.Content{
				Parts: []*genai.Part{
					{
						Text: prompt,
					},
				},
			},
			ResponseMIMEType: "application/json",
			ResponseJsonSchema: map[string]any{
				"type": "array",
				"items": map[string]any{
					"type": "string",
				},
			},
		},
	)
	if err != nil {
		return nil, fmt.Errorf("request gemini: %w", err)
	}
	raw := strings.TrimSpace(resp.Text())

	var nars []string
	if err := json.Unmarshal([]byte(raw), &nars); err != nil {
		return nil, fmt.Errorf("invalid json: %w; raw=%q", err, raw)
	}
	return nars, nil
}
