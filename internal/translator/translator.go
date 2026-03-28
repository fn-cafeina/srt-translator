package translator

import (
	"fmt"
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

		translatedMap, err := t.processChunk(chunk, systemInstruction)
		if err != nil {
			return nil, fmt.Errorf("translation failed for blocks [%s-%s]: %w", chunk[0].ID, chunk[len(chunk)-1].ID, err)
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

func (t *Translator) processChunk(chunk []srt.Block, sysInst string) (map[string]string, error) {
	translatedMap := make(map[string]string)
	var lastErr error

	for attempt := 1; attempt <= t.Config.MaxRetries; attempt++ {
		var missing []srt.Block
		for _, b := range chunk {
			if _, ok := translatedMap[b.ID]; !ok {
				missing = append(missing, b)
			}
		}

		if len(missing) == 0 {
			return translatedMap, nil
		}

		if attempt > 1 {
			fmt.Printf("\r\033[K[retry %d] missing %d blocks...\n", attempt-1, len(missing))
		}

		prompt := t.buildUserPrompt(missing, translatedMap)
		raw, err := t.Client.GenerateText(prompt, sysInst, translationSchema)
		if err != nil {
			lastErr = fmt.Errorf("gemini text generation failed on attempt %d: %w", attempt, err)
			time.Sleep(t.Config.RetryDelay)
			continue
		}

		resMap, err := t.parseResponse(raw)
		if err != nil {
			lastErr = fmt.Errorf("failed to parse gemini response on attempt %d: %w", attempt, err)
			time.Sleep(t.Config.RetryDelay)
			continue
		}

		for _, b := range missing {
			if val, ok := resMap[b.ID]; ok && val != "" {
				translatedMap[b.ID] = val
			}
		}

		if len(translatedMap) < len(chunk) {
			if attempt < t.Config.MaxRetries {
				time.Sleep(t.Config.RetryDelay)
			} else {
				lastErr = fmt.Errorf("missing translations for %d blocks", len(chunk)-len(translatedMap))
			}
		}
	}

	return translatedMap, fmt.Errorf("failed to translate all blocks after %d attempts: %v", t.Config.MaxRetries, lastErr)
}
