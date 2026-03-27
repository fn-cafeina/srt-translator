package translator

import (
	"encoding/json"
	"fmt"

	"github.com/fn-cafeina/srt-translator/internal/gemini"
)

var contextSchema = &gemini.Schema{
	Type: "object",
	Properties: map[string]gemini.Schema{
		"context":        {Type: "string"},
		"sourceLang":     {Type: "string"},
		"targetLangCode": {Type: "string"},
		"cleanName":      {Type: "string"},
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

	raw, err := t.Client.GenerateText(prompt, sysInst, contextSchema)
	if err != nil {
		return nil, err
	}

	var ctx ContextResponse
	if err := json.Unmarshal([]byte(raw), &ctx); err != nil {
		return &ContextResponse{Context: "Unknown", SourceLang: "Unknown", CleanName: filename}, nil
	}

	return &ctx, nil
}
