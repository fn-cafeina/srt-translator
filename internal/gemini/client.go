package gemini

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type GeminiClient struct {
	Config Config
}


func (c *GeminiClient) call(reqBody RequestBody) (string, error) {
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", c.Config.Model, c.Config.ApiKey)
	
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var response Response
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API error (%d): %s", resp.StatusCode, response.Error.Message)
	}

	if len(response.Candidates) == 0 || len(response.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("empty response from API")
	}

	return response.Candidates[0].Content.Parts[0].Text, nil
}



func (c *GeminiClient) GenerateText(prompt, systemInstruction string, schema *Schema) (string, error) {
	config := &GenerationConfig{
		Temperature: c.Config.Temperature,
		TopP:        c.Config.TopP,
	}
	if schema != nil {
		config.ResponseMimeType = "application/json"
		config.ResponseSchema = schema
	}

	req := RequestBody{
		SystemInstruction: &Content{Parts: []Part{{Text: systemInstruction}}},
		Contents: []Content{
			{Role: "user", Parts: []Part{{Text: prompt}}},
		},
		GenerationConfig: config,
	}

	return c.call(req)
}
