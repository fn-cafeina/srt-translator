package env

import (
	"fmt"
	"os"
	"strings"
)

func LoadEnv() error {
	content, err := os.ReadFile(".env")
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("reading .env file failed: %w", err)
	}

	for line := range strings.SplitSeq(string(content), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		key, value, found := strings.Cut(line, "=")
		if !found {
			return fmt.Errorf("malformed line in .env file, missing '=': %q", line)
		}

		if err := os.Setenv(strings.TrimSpace(key), strings.TrimSpace(value)); err != nil {
			return fmt.Errorf("failed to set environment variable %q: %w", key, err)
		}
	}

	return nil
}
