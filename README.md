# SRT Translator (Go CLI)

Professional Go CLI tool to translate `.srt` subtitle files into any language using the Google Gemini API.

## Features
- **Robust JSON Architecture**: Uses structured JSON for AI communications, ensuring 100% reliable mapping.
- **Smart Translation**: Automatically detects movie/series context for high-quality, consistent results.
- **Language Agnostic**: Translate to any target language (Spanish, French, Italian, Japanese, etc.).
- **High Performance**: Optimized for maximum throughput with incremental retry stability.
- **Context Awareness**: Maintains translation consistency throughout the entire file.
- **Robustness**: Automated retries and exponential backoff for API reliability.
- **Minimalist**: Dependency-free core, only using the Gemini API.

## Requirements
- Go 1.21+ (Recommended 1.24+)
- Google AI Studio API Key (Gemini)

## Installation
Build the binary using the provided `Makefile`:
```bash
make build
```

## Usage
1. Copy `.env.example` to `.env` and add your `GEMINI_API_KEY`.
2. Run the translator:
```bash
./bin/srt-translator -input movie.srt -lang "Italian"
```

### Options
- `-input`: Path to input `.srt` file (required).
- `-output`: Custom path for the output file (optional).
- `-lang`: Target language name (default: `Spanish`).
- `-api-key`: Gemini API Key (defaults to `GEMINI_API_KEY` env var).

## Development
- `make build`: Compile the tool.
- `make run ARGS="..."`: Build and run with flags.
- `make clean`: Remove build artifacts.

## TODO
- **Parallel Wave Pipeline**: Implement concurrent pipelines to further double throughput.
- **Adaptive Rate Limiting**: Dynamic delay adjustment based on real-time API feedback.
- **Unit Testing**: Implement comprehensive tests for internal packages.
