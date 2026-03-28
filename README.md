# srt-translator

CLI tool for translating `.srt` files using the Gemini API.

## Requirements
- Go 1.26+
- Google AI Studio API Key (Gemini)
  - Get an API key at: https://aistudio.google.com/app/apikey
  - Model: `gemini-3.1-flash-lite-preview`

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
# Standard
./bin/srt-translator -i movie.srt -l "Spanish"

# With audio context
./bin/srt-translator -i movie.srt -m movie.mkv -l "Spanish"

# Quiet mode
./bin/srt-translator -i movie.srt -q
```

### Options
- `-i` : Input `.srt` file path (required)
- `-m` : Input media file path for audio context (optional)
- `-o` : Output file path (optional)
- `-l` : Target language (default: `Spanish`)
- `-k` : Gemini API Key (optional via flag, required via env)
- `-q` : Quiet mode (optional)

### Output Formatting
If `-o` is omitted, the translated ISO language code is appended to the filename (e.g., `movie.srt` -> `movie_es.srt`).

### Rate Limits
The tool enforces a 4100ms delay between translation chunks to respect the 15 RPM limit of the Gemini Free Tier.
