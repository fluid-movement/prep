package cli

import (
	"fmt"

	"github.com/fluid-movement/prep/internal/update"
)

// cloudScript is the one definition of the cloud environment setup script
// (see the finding "Cloud environment setup script gives a 1:1 prep
// install"). %s is the optional PREP_VERSION line, empty for development
// builds. Every step that can fail ends in || true so a failure never
// blocks session start.
const cloudScript = `#!/bin/bash
%scurl -fsSL https://raw.githubusercontent.com/fluid-movement/prep/main/install.sh | sh || true
echo "claude on PATH: $(command -v claude || echo no)"
"$HOME/.local/bin/prep" setup --harness claude-code || true
echo "prep version: $("$HOME/.local/bin/prep" version 2>&1)"
`

// renderCloudScript renders the script for version v; a development build
// omits the pin so the latest release is installed.
func renderCloudScript(v string) string {
	pin := ""
	if update.IsRelease(v) {
		pin = "export PREP_VERSION=" + v + "\n"
	}
	return fmt.Sprintf(cloudScript, pin)
}

// cmdCloud implements `prep cloud script`: it prints the environment setup
// script for this binary's version to stdout.
func cmdCloud(a *app, args []string) error {
	if len(args) != 1 || args[0] != "script" {
		return usageErr("usage: prep cloud script")
	}
	if !update.IsRelease(Version) {
		fmt.Fprintln(a.errw, "development build: the script installs the latest release")
	}
	a.printf("%s", renderCloudScript(Version))
	return nil
}
