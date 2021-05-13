package main

import (
	"fmt"
	"os"

	"github.com/xenitab/mqtt-log-stdout/pkg/config"
)

func main() {
	err := config.GenerateMarkdown("CLI.md")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to generate documentation: %q\n", err)
		os.Exit(1)
	}

	os.Exit(0)
}
