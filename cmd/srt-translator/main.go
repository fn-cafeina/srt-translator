package main

import (
	"fmt"
	"os"

	"github.com/fn-cafeina/srt-translator/internal/cli"
)

func main() {
	if err := cli.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
