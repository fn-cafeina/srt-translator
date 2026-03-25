package utils

import (
	"os"
	"strings"
)

func LoadEnv() {
	content, err := os.ReadFile(".env")
	if err != nil {
		return
	}
	for line := range strings.SplitSeq(string(content), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, found := strings.Cut(line, "=")
		if found {
			os.Setenv(strings.TrimSpace(key), strings.TrimSpace(value))
		}
	}
}
