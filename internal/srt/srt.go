package srt

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	tagsRegex = regexp.MustCompile(`<[^>]+>|{[^}]+}`)
)

type Block struct {
	ID        string `json:"id"`
	Timestamp string `json:"-"`
	Text      string `json:"text"`
}

func cleanText(text string) string {
	return strings.TrimSpace(tagsRegex.ReplaceAllString(text, ""))
}

func Parse(srtText string) ([]Block, error) {
	srtText = strings.TrimPrefix(srtText, "\ufeff")
	srtText = strings.ReplaceAll(srtText, "\r\n", "\n")

	var blocks []Block
	for raw := range strings.SplitSeq(strings.TrimSpace(srtText), "\n\n") {
		lines := strings.Split(strings.TrimSpace(raw), "\n")
		if len(lines) == 1 && lines[0] == "" {
			continue
		}
		if len(lines) < 3 {
			return nil, fmt.Errorf("malformed SRT block, expected at least 3 lines but got %d: %q", len(lines), raw)
		}

		id := strings.TrimSpace(lines[0])
		id = strings.Map(func(r rune) rune {
			if r < 32 {
				return -1
			}
			return r
		}, id)
		if id == "" {
			return nil, fmt.Errorf("malformed SRT block, missing or invalid ID: %q", raw)
		}

		timestamp := strings.TrimSpace(lines[1])
		if timestamp == "" {
			return nil, fmt.Errorf("malformed SRT block %q, missing timestamp: %q", id, raw)
		}

		text := strings.Join(lines[2:], "\n")
		cleaned := cleanText(text)

		blocks = append(blocks, Block{
			ID:        id,
			Timestamp: timestamp,
			Text:      cleaned, // We allow empty text as it can legitimately appear in some SRTS
		})
	}

	if len(blocks) == 0 {
		return nil, fmt.Errorf("no valid blocks found in SRT file")
	}

	return blocks, nil
}

func Encode(blocks []Block) string {
	var sb strings.Builder
	for _, b := range blocks {
		sb.WriteString(b.ID)
		sb.WriteString("\n")
		sb.WriteString(b.Timestamp)
		sb.WriteString("\n")
		sb.WriteString(b.Text)
		sb.WriteString("\n\n")
	}
	return sb.String()
}
