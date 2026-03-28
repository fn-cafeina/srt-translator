package translator

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/fn-cafeina/srt-translator/internal/srt"
	"github.com/google/generative-ai-go/genai"
)

var translationSchema = &genai.Schema{
	Type: genai.TypeObject,
	Properties: map[string]*genai.Schema{
		"r": {
			Type: genai.TypeArray,
			Items: &genai.Schema{
				Type: genai.TypeObject,
				Properties: map[string]*genai.Schema{
					"i": {Type: genai.TypeString},
					"t": {Type: genai.TypeString},
					"d": {Type: genai.TypeBoolean},
				},
				Required: []string{"i", "t", "d"},
			},
		},
	},
	Required: []string{"r"},
}

func (t *Translator) buildSystemInstruction(lang, context string) string {
	sys := fmt.Sprintf(`ROLE: Expert subtitle translator.
TASK: Translate into natural %s.
FILM CONTEXT: %s

INSTRUCTIONS:
1. Maintain source nuances, gender markers, and cultural context.
2. Use natural, localized phrasing for %s.
3. Keep lines short (≤42 chars, max 2 lines).`, lang, context, lang)

	if t.Config.VideoPath != "" {
		sys += "\n4. Listen to the attached audio snippet corresponding to these subtitles. Identify correct voice genders, emotions, and tones.\n5. If the audio is completely desynchronized from the text timing context, set the 'd' flag to true."
	}
	return sys
}

func (t *Translator) buildUserPrompt(chunk []srt.Block, partialMap map[string]string) string {
	var memoryItems []memoryItem
	memoryItems = append(memoryItems, t.Memory...)

	for id, text := range partialMap {
		memoryItems = append(memoryItems, memoryItem{ID: id, TranslatedText: text})
	}

	maxMem := t.Config.ChunkSize * t.Config.MemoryChunks
	if len(memoryItems) > maxMem {
		memoryItems = memoryItems[len(memoryItems)-maxMem:]
	}

	var sb strings.Builder
	if len(memoryItems) > 0 {
		sb.WriteString("CONTEXT (FOR CONSISTENCY - DO NOT TRANSLATE):\n")
		for _, item := range memoryItems {
			sb.WriteString(fmt.Sprintf("[%s] %s\n", item.ID, item.TranslatedText))
		}
		sb.WriteString("\n")
	}

	type minBlock struct {
		I string `json:"i"`
		T string `json:"t"`
	}
	var minBlocks []minBlock
	for _, b := range chunk {
		minBlocks = append(minBlocks, minBlock{I: b.ID, T: b.Text})
	}

	jsonBytes, _ := json.Marshal(minBlocks)
	sb.WriteString("TARGET TO TRANSLATE:\n")
	sb.WriteString(string(jsonBytes))

	return sb.String()
}

type translationResult struct {
	I string `json:"i"`
	T string `json:"t"`
	D bool   `json:"d"`
}

type translationResponse struct {
	R []translationResult `json:"r"`
}

func (t *Translator) parseResponse(raw string) (map[string]string, bool, error) {
	var resp translationResponse
	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		return nil, false, fmt.Errorf("failed to unmarshal JSON response from Gemini: %w", err)
	}

	mapping := make(map[string]string)
	hasDesync := false
	for _, b := range resp.R {
		mapping[b.I] = b.T
		if b.D {
			hasDesync = true
		}
	}
	return mapping, hasDesync, nil
}
