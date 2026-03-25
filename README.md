# SRT Translator

Minimalist `.srt` translator powered by Gemini. Focused on ease of use and precision.

## Key Features
- **Zero-Config**: No need to select model or source language.
- **Auto-Context**: Intelligent movie/series identification from filename.
- **Smart Naming**: Generates clean, human-readable titles for translated files.
- **Privacy-First**: 100% client-side processing.

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
3. Open `http://localhost:1234`

## Build and Deployment
Generate production build in `dist/`:
```bash
npm run build
```

Push `dist/` to `gh-pages` branch:
```bash
npm run deploy
```

