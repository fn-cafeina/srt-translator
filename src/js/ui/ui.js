import { DOM } from '../utils/dom.js';
import { CONFIG } from '../utils/config.js';

export function uiLog(msg, type = "info") {
  const el = document.createElement("div");
  el.className = `log-${type}`;
  el.innerText = `[${new Date().toLocaleTimeString()}] ${msg}`;
  DOM.log.appendChild(el);
  DOM.log.scrollTop = DOM.log.scrollHeight;
}

export function formatTime(seconds) {
  if (!isFinite(seconds) || seconds < 0) return "--:--";
  const m = Math.floor(seconds / 60).toString().padStart(2, "0");
  const s = Math.floor(seconds % 60).toString().padStart(2, "0");
  return `${m}:${s}`;
}

export function updateProgress(processedChunks, totalChunks, startTime, totalBlocks) {
  const pct = (processedChunks / totalChunks) * 100;
  DOM.progressBar.style.width = `${pct}%`;

  const currentLines = Math.min(processedChunks * CONFIG.CHUNK_SIZE, totalBlocks);
  DOM.progressText.innerText = `${currentLines} / ${totalBlocks} lines`;

  const elapsed = (Date.now() - startTime) / 1000;
  const avg = elapsed / processedChunks;
  const remain = totalChunks - processedChunks;
  DOM.etaText.innerText = `ETA ${formatTime(avg * remain)}`;
}

export function download(fileName, content, isPartial) {
  if (!content.trim()) return;
  const blob = new Blob([content.trim()], { type: "text/plain" });
  const url = URL.createObjectURL(blob);
  const a = document.createElement("a");
  a.download = fileName.replace(".srt", isPartial ? "_ES_PARTIAL.srt" : "_ES.srt");
  a.href = url;
  a.click();
  URL.revokeObjectURL(url);
}
