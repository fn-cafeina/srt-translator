package gemini

import (
	"context"
	"fmt"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

type Config struct {
	ApiKey      string
	Model       string
	Temperature float64
	TopP        float64
}

type GeminiClient struct {
	Config Config
}

func (c *GeminiClient) GenerateText(ctx context.Context, prompt, systemInstruction string, schema *genai.Schema, audioBlob *genai.Blob) (string, error) {
	client, err := genai.NewClient(ctx, option.WithAPIKey(c.Config.ApiKey))
	if err != nil {
		return "", fmt.Errorf("failed to create gemini client: %w", err)
	}
	defer client.Close()

	model := client.GenerativeModel(c.Config.Model)
	model.SetTemperature(float32(c.Config.Temperature))
	model.SetTopP(float32(c.Config.TopP))
	model.SystemInstruction = &genai.Content{Parts: []genai.Part{genai.Text(systemInstruction)}}

	if schema != nil {
		model.ResponseMIMEType = "application/json"
		model.ResponseSchema = schema
	}

	var parts []genai.Part
	if audioBlob != nil {
		parts = append(parts, *audioBlob)
	}
	parts = append(parts, genai.Text(prompt))

	resp, err := model.GenerateContent(ctx, parts...)
	if err != nil {
		return "", fmt.Errorf("gemini api error (model: %s): %w", c.Config.Model, err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("gemini API returned empty candidates response (model: %s)", c.Config.Model)
	}

	part := resp.Candidates[0].Content.Parts[0]
	if txt, ok := part.(genai.Text); ok {
		return string(txt), nil
	}

	return "", fmt.Errorf("unexpected gemini response part type for model %s: %T", c.Config.Model, part)
}
