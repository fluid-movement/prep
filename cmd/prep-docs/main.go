// prep-docs regenerates docs/commands.md from prep help.
package main

import (
	"fmt"
	"os"

	"github.com/fluid-movement/prep/internal/docgen"
)

func main() {
	data, err := docgen.Commands()
	if err == nil {
		err = os.WriteFile("docs/commands.md", data, 0644)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
