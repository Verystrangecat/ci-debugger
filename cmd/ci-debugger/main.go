package main

import (
	"fmt"
	"os"

	"github.com/murataslan1/ci-debugger/internal/cli"
)

// version is set at build time via -ldflags.
var version = "dev"

func main() {
	root := cli.NewRootCmd(version)
	if err := root.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "\033[31mError:\033[0m %v\n", err)
		os.Exit(1)
	}
}
