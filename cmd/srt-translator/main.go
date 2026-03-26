package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fn-cafeina/srt-translator/internal/gemini"
	"github.com/fn-cafeina/srt-translator/internal/srt"
	"github.com/fn-cafeina/srt-translator/internal/translator"
	"github.com/fn-cafeina/srt-translator/internal/utils"
)

func main() {
	utils.LoadEnv()
	inputPath := flag.String("input", "", "Path to the input SRT file")
	flag.StringVar(inputPath, "i", "", "Path to the input SRT file (shorthand)")
	outputPath := flag.String("output", "", "Path to the output SRT file (optional)")
	flag.StringVar(outputPath, "o", "", "Path to the output SRT file (optional) (shorthand)")
	apiKey := flag.String("api-key", "", "Gemini API Key (optional, defaults to GEMINI_API_KEY env var)")
	flag.StringVar(apiKey, "k", "", "Gemini API Key (optional) (shorthand)")
	targetLang := flag.String("lang", "Spanish", "Target language name")
	flag.StringVar(targetLang, "l", "Spanish", "Target language name (shorthand)")
	flag.Parse()

	const (
		defaultModel       = "gemini-3.1-flash-lite-preview"
		defaultChunkSize   = 150
		defaultTemperature = 0.25
	)

	if *inputPath == "" {
		fmt.Println("Error: -input is required")
		flag.Usage()
		os.Exit(1)
	}

	if *apiKey == "" {
		*apiKey = os.Getenv("GEMINI_API_KEY")
	}

	if *apiKey == "" {
		fmt.Println("Error: API Key not found. Use -api-key or set GEMINI_API_KEY environment variable.")
		os.Exit(1)
	}

	content, err := os.ReadFile(*inputPath)
	if err != nil {
		fmt.Printf("Error reading input file: %v\n", err)
		os.Exit(1)
	}

	blocks, err := srt.ParseSRT(string(content))
	if err != nil {
		fmt.Printf("Error parsing SRT: %v\n", err)
		os.Exit(1)
	}

	client := &gemini.GeminiClient{
		Config: gemini.Config{
			ApiKey:      *apiKey,
			Model:       defaultModel,
			Temperature: defaultTemperature,
			TopP:        0.95,
		},
	}

	filename := filepath.Base(*inputPath)
	sample := ""
	for i := 0; i < min(len(blocks), 5); i++ {
		sample += blocks[i].Text + " "
	}

	fmt.Print("Detecting context... ")
	ctxResp, err := client.GenerateContext(filename, sample, *targetLang)
	if err != nil {
		fmt.Printf("failed (using defaults): %v\n", err)
		ctxResp = &gemini.ContextResponse{
			Context:        "General",
			SourceLang:     "Unknown",
			TargetLangCode: "translated",
			CleanName:      filename,
		}
	} else {
		fmt.Printf("Done (%s - %s -> %s)\n", ctxResp.CleanName, ctxResp.SourceLang, ctxResp.TargetLangCode)
	}

	if *outputPath == "" {
		ext := filepath.Ext(*inputPath)
		base := strings.TrimSuffix(*inputPath, ext)
		*outputPath = fmt.Sprintf("%s_%s%s", base, ctxResp.TargetLangCode, ext)
	}

	fmt.Printf("Translating %d blocks from %s to %s\n", len(blocks), *inputPath, *outputPath)

	trans := translator.NewTranslator(translator.Config{
		ChunkSize:    defaultChunkSize,
		MemoryChunks: 1,
		MaxRetries:   4,
		RetryDelay:   7 * time.Second,
		ApiDelay:     4100 * time.Millisecond,
		GeminiConfig: client.Config,
		TargetLang:   *targetLang,
	})

	finalBlocks, err := trans.Translate(blocks, ctxResp.Context, func(processed, total int) {
		fmt.Printf("\rProgress: %d/%d (%.1f%%)", processed, total, float64(processed)/float64(total)*100)
	})
	fmt.Println()

	if err != nil {
		fmt.Printf("Error during translation: %v\n", err)
		os.Exit(1)
	}

	finalContent := srt.Stringify(finalBlocks)
	err = os.WriteFile(*outputPath, []byte(finalContent), 0644)
	if err != nil {
		fmt.Printf("Error writing output file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Success! Translated file saved to %s\n", *outputPath)
}
