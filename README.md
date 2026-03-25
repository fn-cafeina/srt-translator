# SRT Translator

Subtitle translator for `.srt` files into Latin American Spanish using the Gemini API.

## Features
- Automatic metadata discovery (context and source language).
- Normalization of output filenames.
- Sequential chunked translation with previous block persistence.
- Client-side execution in the browser.

## Requirements
- Node.js
- Google AI Studio API Key

## Development
1. Install dependencies:
   ```bash
   npm install
   ```
2. Start development server:
   ```bash
   npm run dev
   ```

## Build
Generate production build in `dist/`:
```bash
npm run build
```

## Deployment
Push `dist/` to `gh-pages` branch:
```bash
npm run deploy
```
