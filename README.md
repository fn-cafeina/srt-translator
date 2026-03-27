# srt-translator

CLI tool for translating `.srt` files using the Gemini API.

## Requirements
- Go 1.26+
- Google AI Studio API Key (Gemini)

## Installation

### Binary
Download pre-built binaries from the [Releases](https://github.com/fn-cafeina/srt-translator/releases) section.

### Go Install
```bash
go install github.com/fn-cafeina/srt-translator/cmd/srt-translator@latest
```

### Source
```bash
make build
```

## Usage

Set the API key via the `GEMINI_API_KEY` environment variable, a `.env` file, or the `-k` flag.

```bash
./bin/srt-translator -i movie.srt -l "Spanish"
```

### Options
- `-i` : Input `.srt` file path (required)
- `-o` : Output file path (optional)
- `-l` : Target language (default: `Spanish`)
- `-k` : Gemini API Key
