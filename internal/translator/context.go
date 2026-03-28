package translator

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/generative-ai-go/genai"
)

var contextSchema = &genai.Schema{
	Type: genai.TypeObject,
	Properties: map[string]*genai.Schema{
		"context":        {Type: genai.TypeString},
		"sourceLang":     {Type: genai.TypeString},
		"targetLangCode": {Type: genai.TypeString},
		"cleanName":      {Type: genai.TypeString},
	},
	Required: []string{"context", "sourceLang", "targetLangCode", "cleanName"},
}

type ContextResponse struct {
	Context        string `json:"context"`
	SourceLang     string `json:"sourceLang"`
	TargetLangCode string `json:"targetLangCode"`
	CleanName      string `json:"cleanName"`
}

func (t *Translator) DetectContext(filename, sample string) (*ContextResponse, error) {
	prompt := fmt.Sprintf("Analyze this SRT file: %s\nSample: %s\nTarget Language: %s", filename, sample, t.Config.TargetLang)
	sysInst := "Expert subtitle analyst. Detect context and suggest a clean filename and ISO code."
	ctxBackground := context.Background()

	raw, err := t.Client.GenerateText(ctxBackground, prompt, sysInst, contextSchema, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to detect context for file %q: %w", filename, err)
	}

	var ctx ContextResponse
	if err := json.Unmarshal([]byte(raw), &ctx); err != nil {
		return nil, fmt.Errorf("failed to parse context response for file %q: %w", filename, err)
	}

	return &ctx, nil
}
