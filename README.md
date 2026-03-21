# SRT Translator

Modular web application to automatically translate `.srt` subtitles to Latin American Spanish using Gemini AI. It runs entirely client-side in the browser.

## Requirements

- Node.js
- Google AI Studio API Key (Gemini)

## Local Development

1. Install packaging dependencies:
   ```bash
   npm install
   ```
2. Start the live development server:
   ```bash
   npm run dev
   ```
   *(Note: `npm run start` and `npm start` also work interchangeably)*
3. Open `http://localhost:1234`

## Build and Deployment

The application uses Parcel to optimally package the resources (HTML, JS, and CSS) and generate them in a static directory for production.

- To generate a production build (`dist/`):
  ```bash
  npm run build
  ```
- To dynamically push the `dist/` folder to your repository's `gh-pages` branch:
  ```bash
  npm run deploy
  ```
