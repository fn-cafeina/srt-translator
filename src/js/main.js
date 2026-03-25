import { DOM } from './utils/dom.js';
import { UI } from './ui/UI.js';
import { SRTEngine } from './core/SRTEngine.js';
import { SubtitleTranslator } from './core/SubtitleTranslator.js';
import { CONFIG } from './utils/config.js';

let translatorInstance = null;

DOM.stopBtn.onclick = () => {
  if (translatorInstance) {
    translatorInstance.stop();
  }
};

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
    const textContext = await file.text();
    const parsedBlocks = SRTEngine.parse(textContext);

    translatorInstance = new SubtitleTranslator(
      key,
      CONFIG.DEFAULT_MODEL,
      contextMsg,
      sourceLang,
    );

    translatorInstance.start(
      parsedBlocks,
      (chunkSrt, processedChunks, totalChunks, startTime, totalBlocks) => {
        DOM.preview.value += chunkSrt;
        DOM.preview.scrollTop = DOM.preview.scrollHeight;
        UI.updateProgress(processedChunks, totalChunks, startTime, totalBlocks);
        UI.log("Chunk complete", "success");
      },
      (finalSrt, wasStopped) => {
        UI.log(
          wasStopped ? "Process stopped." : "Translation finished.",
          "success",
        );
        UI.download(file.name, finalSrt, wasStopped);
        DOM.startBtn.disabled = false;
        DOM.stopBtn.disabled = true;
        translatorInstance = null;
      },
    );
  } catch (error) {
    UI.log(`Critical Error: ${error.message}`, "error");
    DOM.startBtn.disabled = false;
    DOM.stopBtn.disabled = true;
  }
};
