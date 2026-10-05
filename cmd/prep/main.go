// Command prep is a workflow engine and memory for coding agents.
package main

import (
	"os"

	"github.com/fluid-movement/prep/internal/cli"
)

func main() {
	os.Exit(cli.Main(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}
