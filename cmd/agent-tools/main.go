package main

import (
	"os"

	"agent-tools/cmd/agent-tools/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}
