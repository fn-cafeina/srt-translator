package translator

import (
	"time"

	"github.com/fn-cafeina/srt-translator/internal/gemini"
)

type Config struct {
	ChunkSize    int
	MemoryChunks int
	MaxRetries   int
	RetryDelay   time.Duration
	ApiDelay     time.Duration
	GeminiConfig gemini.Config
	TargetLang   string
	VideoPath    string
	Quiet        bool
}

type Translator struct {
	Config Config
	Client *gemini.GeminiClient
	Memory []memoryItem
}

type memoryItem struct {
	ID             string
	TranslatedText string
}
