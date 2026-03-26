# SRT Translator CLI

CLI tool for translating `.srt` files using the Gemini API.

## Features
- **JSON Protocol**: Structural output enforcement via JSON schema.
- **Context Detection**: Automated analysis of subtitle context.
- **Language Support**: Translation to any target language.
- **Throughput Calibration**: 15 RPM / 250k TPM / 500 RPD (150 blocks/chunk, 4.1s delay).
- **Contextual Memory**: Historical buffer for consistency.
- **Retry Logic**: Incremental retries with exponential backoff.
- **Minimalist**: Dependency-free Go core.

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

## Configuration
The tool requires a **Google AI Studio API Key**. You can provide it in three ways (in order of precedence):
1.  **Command flag**: `-api-key "your_key"`
2.  **Environment variable**: Set `GEMINI_API_KEY="your_key"` in your `.bashrc` or `.zshrc`.
3.  **Local file**: Create a `.env` file in the current directory with `GEMINI_API_KEY="your_key"`.

## Usage
Run the translator:
```bash
./bin/srt-translator -input movie.srt -lang "Italian"
```

### Options
- `-i`, `-input`: Path to input `.srt` file (required).
- `-o`, `-output`: Custom path for the output file (optional).
- `-l`, `-lang`: Target language name (default: `Spanish`).
- `-k`, `-api-key`: Gemini API Key (defaults to `GEMINI_API_KEY` env var).

## Development
- `make build`: Compile the tool.
- `make run ARGS="..."`: Build and run with flags.
- `make clean`: Remove build artifacts.
