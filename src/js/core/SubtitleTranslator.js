import { UI } from '../ui/UI.js';
import { CONFIG } from '../utils/config.js';

export class SubtitleTranslator {
  constructor(apiKey, model, contextMsg, sourceLang) {
    this.apiKey = apiKey;
    this.model = model;
    this.contextMsg = contextMsg;
    this.sourceLang = sourceLang;
    this.isRunning = false;
    this.abortController = null;
    this.memoryContext = [];
  }

  getGenderRules() {
    switch (this.sourceLang) {
      case "it":
      case "pt":
        return "SOURCE IS ROMANCE LANGUAGE: Highly gender-inflected. TRUST THE SOURCE. Preserve exact gender markers (masculine/feminine) present in the original text. Use your cultural knowledge to make it sound like a professional Latin American dub.";
      case "fr":
        return "SOURCE IS FRENCH: Maintain the gender markers from the source text where applicable in Spanish. Prioritize natural, culturally accurate translations.";
      case "en":
        return "SOURCE IS ENGLISH: English lacks gender inflection. DO NOT DEFAULT TO MASCULINE IN SPANISH. Paraphrase to absolute neutrality if speaker gender is unknown (e.g. use '¿Quién es?' instead of '¿Quién es él?').";
      case "auto":
      default:
        return "GENDER RULE: If source is gender-neutral, paraphrase to neutrality. If romance language, maintain source gender. Always aim for a natural, non-robotic localization.";
    }
  }

  buildSystemInstruction() {
    const genderRule = this.getGenderRules();
    return `ROLE: Expert subtitle translator and localizer.
TASK: Translate subtitles into natural, fluent Latin American Spanish. Do not be literal if a phrase has a better cultural equivalent.
FILM CONTEXT: ${this.contextMsg || "General film/series context"}
${genderRule}

FORMAT RULES:
1. Return EXACTLY the same [ID] format.
2. Never merge, split, or omit IDs.
3. Keep lines short (≤42 chars). Max 2 lines.
4. Output ONLY the translated blocks. NO introductions, NO markdown, NO conversational text.`;
  }

  buildUserPrompt(chunk, partialMap = {}) {
    let combinedMemory = [...this.memoryContext];
    const extraContextKeys = Object.keys(partialMap);

    if (extraContextKeys.length > 0) {
      const extraContext = extraContextKeys.map((id) => ({
        id,
        spanish: partialMap[id],
      }));
      combinedMemory = [...combinedMemory, ...extraContext].slice(
        -(CONFIG.CHUNK_SIZE * CONFIG.MEMORY_CHUNKS),
      );
    }

    let memoryString = "";
    if (combinedMemory.length > 0) {
      memoryString = `<previous_context>\nDO NOT TRANSLATE THESE. USE FOR CONTEXT ONLY:\n`;
      memoryString +=
        combinedMemory.map((m) => `[${m.id}]\n${m.spanish}`).join("\n\n") +
        `\n</previous_context>\n\n`;
    }

    const blocksToTranslate = chunk
      .map((b) => `[${b.id}]\n${b.text}`)
      .join("\n\n");

    return `${memoryString}<translate_this>\n${blocksToTranslate}\n</translate_this>`;
  }

  async callAPI(prompt) {
    const url = `https://generativelanguage.googleapis.com/v1beta/models/${this.model}:generateContent?key=${this.apiKey}`;

    const response = await fetch(url, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      signal: this.abortController.signal,
      body: JSON.stringify({
        system_instruction: {
          parts: [{ text: this.buildSystemInstruction() }],
        },
        contents: [{ role: "user", parts: [{ text: prompt }] }],
        generationConfig: {
          temperature: CONFIG.TEMPERATURE,
          topP: CONFIG.TOP_P,
        },
      }),
    });

    const data = await response.json();

    if (!response.ok) {
      throw new Error(data.error?.message || `API Error: ${response.status}`);
    }

    return data.candidates[0].content.parts[0].text;
  }

  parseLLMResponse(responseText) {
    const cleanText = responseText
      .replace(/```[a-z]*\n?/gi, "")
      .replace(/```/g, "")
      .replace(/<translate_this>|<\/translate_this>/gi, "")
      .trim();

    const regex = /\[(\d+)\]\s*([\s\S]*?)(?=\n\s*\[\d+\]|$)/g;
    const matches = [...cleanText.matchAll(regex)];

    const map = {};
    matches.forEach((m) => {
        const id = m[1];
        const text = m[2].trim();
        if (text !== "") {
          map[id] = text;
        }
    });

    if (Object.keys(map).length === 0) {
      throw new Error(
        "Format failure: No valid [ID] blocks found in AI response.",
      );
    }

    return map;
  }

  stop() {
    this.isRunning = false;
    if (this.abortController) {
      this.abortController.abort();
    }
  }

  async start(parsedBlocks, onChunkComplete, onFinish) {
    this.isRunning = true;
    this.abortController = new AbortController();

    const totalChunks = Math.ceil(parsedBlocks.length / CONFIG.CHUNK_SIZE);
    let processedChunks = 0;
    const startTime = Date.now();
    let finalSrt = "";

    UI.log(`Loaded ${parsedBlocks.length} blocks.`, "info");
    UI.log(
      `Target: ${this.model} | Lang: ${this.sourceLang.toUpperCase()}`,
      "info",
    );

    for (let i = 0; i < parsedBlocks.length; i += CONFIG.CHUNK_SIZE) {
      if (!this.isRunning) break;

      const chunk = parsedBlocks.slice(i, i + CONFIG.CHUNK_SIZE);
      let translatedMap = {};
      let pendingBlocks = [...chunk];

      UI.log(
        `Translating blocks ${chunk[0].id} to ${chunk[chunk.length - 1].id}`,
      );

      for (let attempt = 1; attempt <= CONFIG.MAX_RETRIES; attempt++) {
        if (!this.isRunning) break;

        try {
          const userPrompt = this.buildUserPrompt(pendingBlocks, translatedMap);
          const rawResponse = await this.callAPI(userPrompt);
          const map = this.parseLLMResponse(rawResponse);

          for (const id in map) {
            if (pendingBlocks.find((b) => b.id === id)) {
              translatedMap[id] = map[id];
            }
          }

          pendingBlocks = chunk.filter((b) => !translatedMap[b.id]);

          if (pendingBlocks.length === 0) {
            break;
          } else {
            UI.log(
              `Attempt ${attempt}: Missing ${pendingBlocks.length} lines. Retrying...`,
              "warn",
            );
            if (attempt < CONFIG.MAX_RETRIES) {
              await new Promise((r) => setTimeout(r, CONFIG.RETRY_DELAY_MS));
            }
          }
        } catch (error) {
          if (error.name === "AbortError") {
            UI.log("Pipeline stopped by user.", "warn");
            break;
          }
          UI.log(`Attempt ${attempt} error: ${error.message}`, "error");
          if (attempt < CONFIG.MAX_RETRIES) {
            await new Promise((r) => setTimeout(r, CONFIG.RETRY_DELAY_MS));
          }
        }
      }

      if (!this.isRunning && Object.keys(translatedMap).length === 0) break;

      let chunkSrt = "";
      const currentTranslatedChunk = [];

      chunk.forEach((block) => {
        const transText = translatedMap[block.id]
          ? translatedMap[block.id]
          : block.text;

        chunkSrt += `${block.id}\n${block.timestamp}\n${transText}\n\n`;
        currentTranslatedChunk.push({ id: block.id, spanish: transText });
      });

      finalSrt += chunkSrt;

      this.memoryContext = [
        ...this.memoryContext,
        ...currentTranslatedChunk,
      ].slice(-(CONFIG.CHUNK_SIZE * CONFIG.MEMORY_CHUNKS));

      processedChunks++;
      onChunkComplete(
        chunkSrt,
        processedChunks,
        totalChunks,
        startTime,
        parsedBlocks.length,
      );

      if (this.isRunning && processedChunks < totalChunks) {
        await new Promise((r) => setTimeout(r, CONFIG.API_DELAY_MS));
      }
    }

    onFinish(finalSrt, !this.isRunning);
  }
}
