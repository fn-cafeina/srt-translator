package srt

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	tagsRegex   = regexp.MustCompile(`<[^>]+>|{[^}]+}`)
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
		if len(lines) < 3 {
			continue
		}

		id := strings.TrimSpace(lines[0])
		id = strings.Map(func(r rune) rune {
			if r < 32 {
				return -1
			}
			return r
		}, id)
		timestamp := lines[1]
		text := strings.Join(lines[2:], "\n")

		cleaned := cleanText(text)
		if id != "" && timestamp != "" && cleaned != "" {
			blocks = append(blocks, Block{
				ID:        id,
				Timestamp: timestamp,
				Text:      cleaned,
			})
		}
	}

	if len(blocks) == 0 {
		return nil, fmt.Errorf("no valid blocks found")
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
