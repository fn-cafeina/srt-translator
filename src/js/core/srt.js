export function cleanText(text) {
  return text.replace(/<[^>]+>|{[^}]+}/g, "").trim();
}

export function parseSRT(srtText) {
  const blocks = srtText
    .replace(/\r\n/g, "\n")
    .trim()
    .split(/\n\s*\n/);

  return blocks
    .map((block) => {
      const lines = block.split("\n");
      const id = lines[0];
      const timestamp = lines[1];
      const rawText = lines.slice(2).join("\n");
      return { id, timestamp, text: cleanText(rawText) };
    })
    .filter((b) => b.id && b.timestamp && b.text !== "");
}
