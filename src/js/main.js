import { DOM } from './utils/dom.js';
import { CONFIG } from './utils/config.js';
import { parseSRT } from './core/srt.js';
import { translateSubtitle, stopTranslation } from './core/translator.js';
import { uiLog, updateProgress, download } from './ui/ui.js';

DOM.stopBtn.onclick = () => stopTranslation();

DOM.startBtn.onclick = async () => {
  const key = DOM.apiKey.value.trim();
  const sourceLang = DOM.sourceLang.value;
  const contextMsg = DOM.context.value;
  const file = DOM.file.files[0];

  if (!key || !file) {
    alert("API Key and SRT File are required.");
    return;
  }

  DOM.startBtn.disabled = true;
  DOM.stopBtn.disabled = false;
  DOM.log.innerHTML = "";
  DOM.preview.value = "";

  try {
    const text = await file.text();
    const blocks = parseSRT(text);

    uiLog(`Loaded ${blocks.length} blocks.`, "info");
    uiLog(`Target: ${CONFIG.DEFAULT_MODEL} | Lang: ${sourceLang.toUpperCase()}`, "info");

    const params = {
      apiKey: key,
      model: CONFIG.DEFAULT_MODEL,
      sourceLang,
      contextMsg,
      parsedBlocks: blocks
    };

    await translateSubtitle(
      params,
      (chunkSrt, processedChunks, totalChunks, startTime) => {
        DOM.preview.value += chunkSrt;
        DOM.preview.scrollTop = DOM.preview.scrollHeight;
        updateProgress(processedChunks, totalChunks, startTime, blocks.length);
        uiLog("Chunk complete", "success");
      },
      (finalSrt, wasStopped) => {
        uiLog(wasStopped ? "Process stopped." : "Translation finished.", "success");
        download(file.name, finalSrt, wasStopped);
        DOM.startBtn.disabled = false;
        DOM.stopBtn.disabled = true;
      }
    );
  } catch (error) {
    uiLog(`Critical Error: ${error.message}`, "error");
    DOM.startBtn.disabled = false;
    DOM.stopBtn.disabled = true;
  }
};
