package main

import (
	"os"

	"tars/cmd/tars/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}
