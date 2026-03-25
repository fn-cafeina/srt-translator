# SRT Translator (Go CLI)

Professional Go CLI tool to translate `.srt` subtitle files into any language using the Gemini 3.1 Flash Preview model.

## Features
- **Robust JSON Architecture**: Uses structured JSON for AI communications, ensuring 100% reliable mapping.
- **Smart Translation**: Automatically detects movie/series context for high-quality, consistent results.
- **Language Agnostic**: Translate to any target language (Spanish, French, Italian, Japanese, etc.).
- **High Performance**: Optimized for 15 RPM / 250k TPM / 500 RPD (150 blocks per chunk, 4.1s delay).
- **Context Awareness**: Maintains translation consistency throughout the entire file.
- **Robustness**: Automated retries and exponential backoff for API reliability.
- **Minimalist**: Dependency-free core, only using the Gemini API.

## Requirements
- Go 1.21+ (Recommended 1.24+)
- Google AI Studio API Key (Gemini)

## Installation

### Binary Downloads
Pre-built binaries for Windows, Linux, and macOS are available in the [Releases](https://github.com/fn-cafeina/srt-translator/releases) section.

### Global Installation
Install the binary directly using Go:
```bash
go install github.com/fn-cafeina/srt-translator/cmd/srt-translator@latest
```

### Build from Source
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
