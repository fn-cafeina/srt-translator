export const SRTEngine = {
  cleanText(text) {
    return text.replace(/<[^>]+>|{[^}]+}/g, "").trim();
  },

  parse(srtText) {
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

        const cleanedText = this.cleanText(rawText);

        return { id, timestamp, text: cleanedText };
      })
      .filter((b) => b.id && b.timestamp && b.text !== "");
  },
};
