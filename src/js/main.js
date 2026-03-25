import { DOM } from './dom.js';
import { CONFIG } from './config.js';
import { parseSRT } from './srt.js';
import { translateSubtitle, stopTranslation, generateContext } from './translator.js';
import { uiLog, updateProgress, download } from './ui.js';

DOM.stopBtn.onclick = () => stopTranslation();

DOM.startBtn.onclick = async () => {
  const key = DOM.apiKey.value.trim();
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

    uiLog("Analyzing file and context...", "info");
    const { context, lang, cleanName } = await generateContext(key, CONFIG.DEFAULT_MODEL, file.name, text);
    uiLog(`✓ Context: ${context.substring(0, 60)}...`, "success");
    uiLog(`✓ Source: ${lang} | Blocks: ${blocks.length}`, "success");

    const params = {
      apiKey: key,
      model: CONFIG.DEFAULT_MODEL,
      contextMsg: context,
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
        const downloadName = cleanName ? `${cleanName}.srt` : file.name;
        download(downloadName, finalSrt, wasStopped);
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
