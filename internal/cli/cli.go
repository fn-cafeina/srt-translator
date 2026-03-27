package cli

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fn-cafeina/srt-translator/internal/env"
	"github.com/fn-cafeina/srt-translator/internal/gemini"
	"github.com/fn-cafeina/srt-translator/internal/srt"
	"github.com/fn-cafeina/srt-translator/internal/translator"
)

const (
	defaultModel       = "gemini-3.1-flash-lite-preview"
	defaultChunkSize   = 150
	defaultTemperature = 0.25
)

func Run() error {
	env.LoadEnv()
	inputPath := flag.String("i", "", "Path to the input SRT file")
	outputPath := flag.String("o", "", "Path to the output SRT file (optional)")
	apiKey := flag.String("k", "", "Gemini API Key (optional, defaults to GEMINI_API_KEY env var)")
	targetLang := flag.String("l", "Spanish", "Target language name")
	flag.Parse()

	if *inputPath == "" {
		flag.Usage()
		return fmt.Errorf("-i (input) is required")
	}

	if *apiKey == "" {
		*apiKey = os.Getenv("GEMINI_API_KEY")
	}

	if *apiKey == "" {
		return fmt.Errorf("API Key not found. Use -k or set GEMINI_API_KEY environment variable")
	}

	content, err := os.ReadFile(*inputPath)
	if err != nil {
		return fmt.Errorf("reading input file: %w", err)
	}

	blocks, err := srt.Parse(string(content))
	if err != nil {
		return fmt.Errorf("parsing SRT: %w", err)
	}

	client := &gemini.GeminiClient{
		Config: gemini.Config{
			ApiKey:      *apiKey,
			Model:       defaultModel,
			Temperature: defaultTemperature,
			TopP:        0.95,
		},
	}

	trans := translator.NewTranslator(translator.Config{
		ChunkSize:    defaultChunkSize,
		MemoryChunks: 1,
		MaxRetries:   4,
		RetryDelay:   7 * time.Second,
		ApiDelay:     4100 * time.Millisecond,
		GeminiConfig: client.Config,
		TargetLang:   *targetLang,
	})

	filename := filepath.Base(*inputPath)
	sample := ""
	for i := 0; i < min(len(blocks), 5); i++ {
		sample += blocks[i].Text + " "
	}

	fmt.Print("detecting context... ")
	ctxResp, err := trans.DetectContext(filename, sample)
	if err != nil {
		fmt.Printf("failed\n\n")
		ctxResp = &translator.ContextResponse{
			Context:        "General",
			SourceLang:     "Unknown",
			TargetLangCode: "translated",
			CleanName:      filename,
		}
	} else {
		fmt.Printf("done (%s -> %s)\n\n", strings.ToLower(ctxResp.SourceLang), strings.ToLower(ctxResp.TargetLangCode))
	}

	if *outputPath == "" {
		ext := filepath.Ext(*inputPath)
		base := strings.TrimSuffix(*inputPath, ext)
		*outputPath = fmt.Sprintf("%s_%s%s", base, ctxResp.TargetLangCode, ext)
	}

	fmt.Printf("translating %d blocks...\n", len(blocks))

	finalBlocks, err := trans.Translate(blocks, ctxResp.Context, func(processed, total int) {
		fmt.Printf("\r\033[Kprogress: %d/%d (%.1f%%)", processed, total, float64(processed)/float64(total)*100)
	})
	fmt.Println()

	if err != nil {
		return fmt.Errorf("during translation: %w", err)
	}

	finalContent := srt.Encode(finalBlocks)
	if err := os.WriteFile(*outputPath, []byte(finalContent), 0644); err != nil {
		return fmt.Errorf("writing output file: %w", err)
	}

	fmt.Printf("\nsuccess. saved to %s\n", *outputPath)
	return nil
}
