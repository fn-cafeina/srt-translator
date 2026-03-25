package translator

import (
	"fmt"
	"strings"
	"time"

	"github.com/fn-cafeina/srt-translator/internal/gemini"
	"github.com/fn-cafeina/srt-translator/internal/srt"
)

func NewTranslator(cfg Config) *Translator {
	return &Translator{
		Config: cfg,
		Client: &gemini.GeminiClient{Config: cfg.GeminiConfig},
	}
}

func (t *Translator) Translate(blocks []srt.Block, contextMsg string, onProgress func(int, int)) ([]srt.Block, error) {
	systemInstruction := t.buildSystemInstruction(t.Config.TargetLang, contextMsg)
	var finalBlocks []srt.Block
	total := len(blocks)
	processed := 0

	for i := 0; i < total; i += t.Config.ChunkSize {
		end := min(i+t.Config.ChunkSize, total)
		chunk := blocks[i:end]
		translatedMap := make(map[string]string)
		
		var err error
		for attempt := 1; attempt <= t.Config.MaxRetries; attempt++ {
			var missing []srt.Block
			for _, b := range chunk {
				if _, ok := translatedMap[b.ID]; !ok {
					missing = append(missing, b)
				}
			}

			if len(missing) == 0 {
				err = nil
				break
			}

			if attempt > 1 {
				var ids []string
				for _, b := range missing {
					ids = append(ids, b.ID)
				}
				fmt.Printf("\n[Retry %d] Missing IDs: %s. Retrying...", attempt-1, strings.Join(ids, ", "))
			}

			prompt := t.buildUserPrompt(missing, translatedMap)
			var raw string
			raw, err = t.Client.Translate(prompt, systemInstruction, translationSchema)
			if err == nil {
				resMap := t.parseResponse(raw)
				for _, b := range missing {
					if val, ok := resMap[b.ID]; ok && val != "" {
						translatedMap[b.ID] = val
					}
				}
			}

			if attempt < t.Config.MaxRetries && len(translatedMap) < len(chunk) {
				time.Sleep(t.Config.RetryDelay)
			}
		}

		if err != nil {
			return nil, fmt.Errorf("translation failed after %d attempts: %w", t.Config.MaxRetries, err)
		}

		for _, b := range chunk {
			text := translatedMap[b.ID]
			if text == "" {
				text = b.Text
			}
			newBlock := b
			newBlock.Text = text
			finalBlocks = append(finalBlocks, newBlock)
			t.Memory = append(t.Memory, memoryItem{ID: b.ID, TranslatedText: text})
		}

		processed += len(chunk)
		if onProgress != nil {
			onProgress(processed, total)
		}

		if processed < total {
			time.Sleep(t.Config.ApiDelay)
		}
	}

	return finalBlocks, nil
}
