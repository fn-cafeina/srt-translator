import { CONFIG } from './config.js';

let isRunning = false;
let abortController = null;
let memoryContext = [];

export function stopTranslation() {
  isRunning = false;
  if (abortController) abortController.abort();
}

export async function generateContext(apiKey, model, filename, sampleText) {
  const prompt = `Based on filename "${filename}" and these lines: "${sampleText.substring(0, 300)}", 
  provide: 1. Brief context (min words, latin only), 2. Detected source language name. 
  Return JSON ONLY: {"context": "...", "lang": "..."}`;
  
  const url = `https://generativelanguage.googleapis.com/v1beta/models/${model}:generateContent?key=${apiKey}`;
  const response = await fetch(url, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({
      contents: [{ role: "user", parts: [{ text: prompt }] }],
      generationConfig: { 
        temperature: 0.2,
        response_mime_type: "application/json"
      }
    }),
  });
  
  const data = await response.json();
  if (!response.ok) throw new Error(data.error?.message || `API Error: ${response.status}`);
  
  try {
    return JSON.parse(data.candidates[0].content.parts[0].text);
  } catch (e) {
    return { context: data.candidates[0].content.parts[0].text, lang: "Unknown" };
  }
}

function getGenderRules(sourceLang) {
  switch (sourceLang) {
    case "it":
    case "pt":
      return "SOURCE IS ROMANCE LANGUAGE: Preserve exact gender markers. Use cultural knowledge for natural Latin American Spanish.";
    case "fr":
      return "SOURCE IS FRENCH: Maintain gender markers. Prioritize natural localization.";
    case "en":
      return "SOURCE IS ENGLISH: DO NOT DEFAULT TO MASCULINE. Paraphrase to neutrality if gender is unknown.";
    default:
      return "GENDER RULE: Paraphrase to neutrality if source is neutral. Maintain romance gender. Aim for natural localization.";
  }
}

function buildSystemInstruction(sourceLang, contextMsg) {
  const genderRule = getGenderRules(sourceLang);
  return `ROLE: Expert subtitle translator.
TASK: Translate into natural Latin American Spanish.
FILM CONTEXT: ${contextMsg || "General"}
${genderRule}

FORMAT:
1. Return EXACTLY the same [ID] format.
2. Never merge, split, or omit IDs.
3. Keep lines short (≤42 chars). Max 2 lines.
4. Output ONLY the translated blocks. NO introductions, NO markdown.`;
}

function buildUserPrompt(chunk, partialMap = {}) {
  let combinedMemory = [...memoryContext];
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
    memoryString = `<previous_context>\nDO NOT TRANSLATE:\n`;
    memoryString +=
      combinedMemory.map((m) => `[${m.id}]\n${m.spanish}`).join("\n\n") +
      `\n</previous_context>\n\n`;
  }

  const blocks = chunk.map((b) => `[${b.id}]\n${b.text}`).join("\n\n");
  return `${memoryString}<translate_this>\n${blocks}\n</translate_this>`;
}

async function callAPI(apiKey, model, prompt, systemInstruction, signal) {
  const url = `https://generativelanguage.googleapis.com/v1beta/models/${model}:generateContent?key=${apiKey}`;
  const response = await fetch(url, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    signal,
    body: JSON.stringify({
      system_instruction: { parts: [{ text: systemInstruction }] },
      contents: [{ role: "user", parts: [{ text: prompt }] }],
      generationConfig: {
        temperature: CONFIG.TEMPERATURE,
        topP: CONFIG.TOP_P,
      },
    }),
  });

  const data = await response.json();
  if (!response.ok) throw new Error(data.error?.message || `API Error: ${response.status}`);
  return data.candidates[0].content.parts[0].text;
}

function parseResponse(responseText) {
  const clean = responseText.replace(/```[a-z]*\n?|```|<translate_this>|<\/translate_this>/gi, "").trim();
  const regex = /\[(\d+)\]\s*([\s\S]*?)(?=\n\s*\[\d+\]|$)/g;
  const matches = [...clean.matchAll(regex)];
  
  const map = {};
  matches.forEach((m) => {
    const id = m[1];
    const text = m[2].trim();
    if (text) map[id] = text;
  });

  if (Object.keys(map).length === 0) throw new Error("Format failure: No valid [ID] found.");
  return map;
}

export async function translateSubtitle(params, onChunk, onFinish) {
  const { apiKey, model, sourceLang, contextMsg, parsedBlocks } = params;
  isRunning = true;
  abortController = new AbortController();
  memoryContext = [];

  const totalChunks = Math.ceil(parsedBlocks.length / CONFIG.CHUNK_SIZE);
  let processedChunks = 0;
  const startTime = Date.now();
  let finalSrt = "";

  const systemInstruction = buildSystemInstruction(sourceLang, contextMsg);

  for (let i = 0; i < parsedBlocks.length; i += CONFIG.CHUNK_SIZE) {
    if (!isRunning) break;

    const chunk = parsedBlocks.slice(i, i + CONFIG.CHUNK_SIZE);
    let translatedMap = {};
    let pendingBlocks = [...chunk];

    for (let attempt = 1; attempt <= CONFIG.MAX_RETRIES; attempt++) {
      if (!isRunning) break;
      try {
        const prompt = buildUserPrompt(pendingBlocks, translatedMap);
        const raw = await callAPI(apiKey, model, prompt, systemInstruction, abortController.signal);
        const map = parseResponse(raw);

        for (const id in map) {
          if (pendingBlocks.find((b) => b.id === id)) translatedMap[id] = map[id];
        }

        pendingBlocks = chunk.filter((b) => !translatedMap[b.id]);
        if (pendingBlocks.length === 0) break;
        
        if (attempt < CONFIG.MAX_RETRIES) {
          await new Promise((r) => setTimeout(r, CONFIG.RETRY_DELAY_MS));
        }
      } catch (error) {
        if (error.name === "AbortError") break;
        if (attempt < CONFIG.MAX_RETRIES) {
          await new Promise((r) => setTimeout(r, CONFIG.RETRY_DELAY_MS));
        } else {
          throw error;
        }
      }
    }

    if (!isRunning && Object.keys(translatedMap).length === 0) break;

    let chunkSrt = "";
    const currentTranslated = [];

    chunk.forEach((block) => {
      const text = translatedMap[block.id] || block.text;
      chunkSrt += `${block.id}\n${block.timestamp}\n${text}\n\n`;
      currentTranslated.push({ id: block.id, spanish: text });
    });

    finalSrt += chunkSrt;
    memoryContext = [...memoryContext, ...currentTranslated].slice(-(CONFIG.CHUNK_SIZE * CONFIG.MEMORY_CHUNKS));
    processedChunks++;
    
    onChunk(chunkSrt, processedChunks, totalChunks, startTime);

    if (isRunning && processedChunks < totalChunks) {
      await new Promise((r) => setTimeout(r, CONFIG.API_DELAY_MS));
    }
  }

  onFinish(finalSrt, !isRunning);
}
