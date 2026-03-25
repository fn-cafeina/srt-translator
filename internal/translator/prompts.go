package translator

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/fn-cafeina/srt-translator/internal/gemini"
	"github.com/fn-cafeina/srt-translator/internal/srt"
)

var translationSchema = &gemini.Schema{
	Type: "array",
	Items: &gemini.Schema{
		Type: "object",
		Properties: map[string]gemini.Schema{
			"id":   {Type: "string"},
			"text": {Type: "string"},
		},
		Required: []string{"id", "text"},
	},
}

type jsonBlock struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

func (t *Translator) buildSystemInstruction(lang, context string) string {
	return fmt.Sprintf(`ROLE: Expert subtitle translator.
TASK: Translate into natural %s.
FILM CONTEXT: %s

INSTRUCTIONS:
1. Maintain source nuances, gender markers, and cultural context.
2. Use natural, localized phrasing for %s.
3. Keep lines short (≤42 chars, max 2 lines).

INSTRUCTIONS:
1. Maintain source nuances, gender markers, and cultural context.
2. Use natural, localized phrasing for %s.
3. Keep lines short (≤42 chars, max 2 lines).`, lang, context, lang)
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

	var toTranslate []jsonBlock
	for _, b := range chunk {
		toTranslate = append(toTranslate, jsonBlock{ID: b.ID, Text: b.Text})
	}

	jsonBytes, _ := json.Marshal(toTranslate)
	sb.WriteString("TARGET TO TRANSLATE:\n")
	sb.WriteString(string(jsonBytes))

	return sb.String()
}

func (t *Translator) parseResponse(raw string) map[string]string {
	var translated []jsonBlock
	if err := json.Unmarshal([]byte(raw), &translated); err != nil {
		return nil
	}

	mapping := make(map[string]string)
	for _, b := range translated {
		mapping[b.ID] = b.Text
	}
	return mapping
}
