package main

import (
	"os"

	"gitlab.com/distributed_lab/acs/github-module/internal/cli"
)

func main() {
	if !cli.Run(os.Args) {
		os.Exit(1)
	}
}
