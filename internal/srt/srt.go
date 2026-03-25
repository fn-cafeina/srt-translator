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
	ID        string
	Timestamp string
	Text      string
}

func CleanText(text string) string {
	return strings.TrimSpace(tagsRegex.ReplaceAllString(text, ""))
}

func ParseSRT(srtText string) ([]Block, error) {
	// Strip UTF-8 BOM if present
	srtText = strings.TrimPrefix(srtText, "\ufeff")
	srtText = strings.ReplaceAll(srtText, "\r\n", "\n")
	
	var blocks []Block
	for raw := range strings.SplitSeq(strings.TrimSpace(srtText), "\n\n") {
		lines := strings.Split(strings.TrimSpace(raw), "\n")
		if len(lines) < 3 {
			continue
		}

		id := strings.TrimSpace(lines[0])
		// Remove any hidden control chars from ID
		id = strings.Map(func(r rune) rune {
			if r < 32 {
				return -1
			}
			return r
		}, id)
		timestamp := lines[1]
		text := strings.Join(lines[2:], "\n")

		cleaned := CleanText(text)
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

func Stringify(blocks []Block) string {
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
