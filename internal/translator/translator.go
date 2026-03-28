package translator

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/fn-cafeina/srt-translator/internal/audio"
	"github.com/fn-cafeina/srt-translator/internal/gemini"
	"github.com/fn-cafeina/srt-translator/internal/srt"
	"github.com/google/generative-ai-go/genai"
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
	ctx := context.Background()

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

		if attempt > 1 && !t.Config.Quiet {
			fmt.Printf("\r\033[K[retry %d] missing %d blocks...\n", attempt-1, len(missing))
		}

		prompt, err := t.buildUserPrompt(missing, translatedMap)
		if err != nil {
			return nil, fmt.Errorf("failed assembling structural prompt context for missing %d fragments: %w", len(missing), err)
		}

		var audioBlob *genai.Blob
		if t.Config.VideoPath != "" {
			if !audio.HasFFmpeg() {
				return nil, fmt.Errorf("ffmpeg binary not found in system PATH. Required for video/audio extraction")
			}
			sliced, err := audio.SliceChunk(t.Config.VideoPath, chunk[0].Timestamp, chunk[len(chunk)-1].Timestamp)
			if err != nil {
				return nil, fmt.Errorf("audio slicing failed: %w", err)
			}

			audioData, err := os.ReadFile(sliced)
			os.Remove(sliced)
			if err != nil {
				return nil, fmt.Errorf("failed to read sliced audio: %w", err)
			}
			audioBlob = &genai.Blob{
				MIMEType: "audio/mp3",
				Data:     audioData,
			}
		}

		raw, err := t.Client.GenerateText(ctx, prompt, sysInst, translationSchema, audioBlob)
		if err != nil {
			lastErr = fmt.Errorf("gemini text generation failed on attempt %d: %w", attempt, err)
			time.Sleep(t.Config.RetryDelay)
			continue
		}

		resMap, hasDesync, err := t.parseResponse(raw)
		if err != nil {
			lastErr = fmt.Errorf("failed to parse gemini response on attempt %d: %w", attempt, err)
			time.Sleep(t.Config.RetryDelay)
			continue
		}

		if hasDesync {
			return nil, fmt.Errorf("AI detected critical audio/SRT desync at blocks [%s-%s]. Aborting translation globally", chunk[0].ID, chunk[len(chunk)-1].ID)
		}

		for _, b := range missing {
			if val, ok := resMap[b.ID]; ok && val != "" {
				translatedMap[b.ID] = val
			}
		}

		lastErr = fmt.Errorf("missing translations for %d blocks", len(chunk)-len(translatedMap))
		if attempt < t.Config.MaxRetries {
			time.Sleep(t.Config.RetryDelay)
		}
	}

	return translatedMap, fmt.Errorf("failed to translate all blocks after %d attempts: %v", t.Config.MaxRetries, lastErr)
}
