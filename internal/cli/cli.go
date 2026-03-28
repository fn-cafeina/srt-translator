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

type Config struct {
	InputPath  string
	OutputPath string
	ApiKey     string
	TargetLang string
	VideoPath  string
	Quiet      bool
}

func parseConfig() (Config, error) {
	if err := env.LoadEnv(); err != nil {
		return Config{}, fmt.Errorf("loading environment variables: %w", err)
	}

	inputPath := flag.String("i", "", "Path to the input SRT file")
	outputPath := flag.String("o", "", "Path to the output SRT file (optional)")
	apiKey := flag.String("k", "", "Gemini API Key (optional, defaults to GEMINI_API_KEY env var)")
	targetLang := flag.String("l", "Spanish", "Target language name")
	mediaPath := flag.String("m", "", "Path to the input video or audio file (optional)")
	quiet := flag.Bool("q", false, "Quiet mode (suppress all terminal console output)")
	flag.Parse()

	if *inputPath == "" {
		flag.Usage()
		return Config{}, fmt.Errorf("mandatory flag -i (input path) is missing")
	}

	finalApiKey := *apiKey
	if finalApiKey == "" {
		finalApiKey = os.Getenv("GEMINI_API_KEY")
	}

	if finalApiKey == "" {
		return Config{}, fmt.Errorf("gemini api key is missing")
	}

	return Config{
		InputPath:  *inputPath,
		OutputPath: *outputPath,
		ApiKey:     finalApiKey,
		TargetLang: *targetLang,
		VideoPath:  *mediaPath,
		Quiet:      *quiet,
	}, nil
}

func setupTranslator(cfg Config) *translator.Translator {
	client := &gemini.GeminiClient{
		Config: gemini.Config{
			ApiKey:      cfg.ApiKey,
			Model:       defaultModel,
			Temperature: defaultTemperature,
			TopP:        0.95,
		},
	}

	return translator.NewTranslator(translator.Config{
		ChunkSize:    defaultChunkSize,
		MemoryChunks: 1,
		MaxRetries:   4,
		RetryDelay:   7 * time.Second,
		ApiDelay:     4100 * time.Millisecond,
		GeminiConfig: client.Config,
		TargetLang:   cfg.TargetLang,
		VideoPath:    cfg.VideoPath,
		Quiet:        cfg.Quiet,
	})
}

func detectContext(trans *translator.Translator, filename string, blocks []srt.Block, quiet bool) (*translator.ContextResponse, error) {
	sample := ""
	for i := 0; i < min(len(blocks), 5); i++ {
		sample += blocks[i].Text + " "
	}

	if !quiet {
		fmt.Print("detecting context... ")
	}
	ctxResp, err := trans.DetectContext(filename, sample)
	if err != nil {
		if !quiet {
			fmt.Printf("failed\n\n")
		}
		return nil, fmt.Errorf("context detection failed: %w", err)
	}

	if !quiet {
		fmt.Printf("done (%s -> %s)\n\n", strings.ToLower(ctxResp.SourceLang), strings.ToLower(ctxResp.TargetLangCode))
	}
	return ctxResp, nil
}

func Run() error {
	cfg, err := parseConfig()
	if err != nil {
		return err
	}

	content, err := os.ReadFile(cfg.InputPath)
	if err != nil {
		return fmt.Errorf("failed to read input file %q: %w", cfg.InputPath, err)
	}

	blocks, err := srt.Parse(string(content))
	if err != nil {
		return fmt.Errorf("failed to parse SRT file %q: %w", cfg.InputPath, err)
	}

	trans := setupTranslator(cfg)
	filename := filepath.Base(cfg.InputPath)

	ctxResp, err := detectContext(trans, filename, blocks, cfg.Quiet)
	if err != nil {
		return err
	}

	if cfg.OutputPath == "" {
		ext := filepath.Ext(cfg.InputPath)
		base := strings.TrimSuffix(cfg.InputPath, ext)
		cfg.OutputPath = fmt.Sprintf("%s_%s%s", base, ctxResp.TargetLangCode, ext)
	}

	if !cfg.Quiet {
		fmt.Printf("translating %d blocks...\n", len(blocks))
	}

	finalBlocks, err := trans.Translate(blocks, ctxResp.Context, func(processed, total int) {
		if !cfg.Quiet {
			fmt.Printf("\r\033[Kprogress: %d/%d (%.1f%%)", processed, total, float64(processed)/float64(total)*100)
		}
	})

	if !cfg.Quiet {
		fmt.Println()
	}

	if err != nil {
		return fmt.Errorf("translation failed for %q: %w", cfg.InputPath, err)
	}

	finalContent := srt.Encode(finalBlocks)
	if err := os.WriteFile(cfg.OutputPath, []byte(finalContent), 0644); err != nil {
		return fmt.Errorf("failed to write output file to %q: %w", cfg.OutputPath, err)
	}

	if !cfg.Quiet {
		fmt.Printf("\nsuccess. saved to %s\n", cfg.OutputPath)
	}
	return nil
}
