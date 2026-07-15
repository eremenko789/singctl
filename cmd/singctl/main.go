package main

import (
	"os"

	"github.com/eremenko789/singctl/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}
