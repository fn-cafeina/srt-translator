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
./bin/srt-translator -i movie.srt -l "Spanish"
```

### Options
- `-i` : Input `.srt` file path (required)
- `-o` : Output file path (optional)
- `-l` : Target language (default: `Spanish`)
- `-k` : Gemini API Key (optional via flag, required via env)

### Output Formatting
If the `-o` flag is omitted, the tool dynamically appends the translated ISO language code to the output filename (e.g., `movie.srt` -> `movie_es.srt`). This matches standard nomenclature required by media servers like Plex, Jellyfin, and Emby.

### Rate Limits
The tool injects a `4100ms` delay between translation chunks. This correctly manages the `15 RPM` (Requests Per Minute) limitation forced by the Gemini Free Tier, ensuring structural integrity overhead without API drops on large file inputs.
