// Package claudecode integrates prep with Claude Code: it installs the prep
// plugin, served from the prep repository as its own marketplace, at user
// scope through Claude Code's plugin commands.
package claudecode

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/fluid-movement/prep/internal/setup"
	"github.com/fluid-movement/prep/internal/update"
)

const (
	marketplace = "prep"
	pluginID    = "prep@prep"
	// SourceEnv names a marketplace source to use instead of the release
	// tag, such as a prep checkout for development builds.
	SourceEnv = "PREP_PLUGIN_SOURCE"
)

func init() { setup.Register(&Harness{}) }

// Harness is the Claude Code integration.
type Harness struct {
	// Run executes the claude CLI; nil runs the real one. Tests replace it.
	Run func(args ...string) (string, error)
}

func (h *Harness) Name() string  { return "claude-code" }
func (h *Harness) Title() string { return "Claude Code" }

func (h *Harness) Detect() bool {
	if h.Run != nil {
		return true
	}
	_, err := exec.LookPath("claude")
	return err == nil
}

func (h *Harness) run(args ...string) (string, error) {
	if h.Run != nil {
		return h.Run(args...)
	}
	out, err := exec.Command("claude", args...).CombinedOutput()
	if err != nil {
		return string(out), fmt.Errorf("claude %s: %v: %s", strings.Join(args, " "), err, strings.TrimSpace(lastLine(string(out))))
	}
	return string(out), nil
}

type plugin struct {
	ID      string `json:"id"`
	Version string `json:"version"`
	Scope   string `json:"scope"`
}

// plugins reads claude plugin list.
func (h *Harness) plugins() ([]plugin, error) {
	out, err := h.run("plugin", "list", "--json")
	if err != nil {
		return nil, err
	}
	var ps []plugin
	if err := json.Unmarshal([]byte(jsonPart(out)), &ps); err != nil {
		return nil, fmt.Errorf("reading claude plugin list: %v", err)
	}
	return ps, nil
}

// Installed reads the plugin's version from claude plugin list.
func (h *Harness) Installed() (string, bool, error) {
	plugins, err := h.plugins()
	if err != nil {
		return "", false, err
	}
	for _, p := range plugins {
		if p.ID == pluginID && p.Scope == "user" {
			return "v" + strings.TrimPrefix(p.Version, "v"), true, nil
		}
	}
	return "", false, nil
}

// Source is the marketplace source for a binary version: the release tag
// of the prep repository, or PREP_PLUGIN_SOURCE.
func Source(version string) (string, error) {
	if s := os.Getenv(SourceEnv); s != "" {
		return s, nil
	}
	if !update.IsRelease(version) {
		return "", fmt.Errorf("this is a development build (%s): set %s to a prep checkout, or install a release", version, SourceEnv)
	}
	return "fluid-movement/prep#" + version, nil
}

// Install pins the marketplace to the source for version and installs the
// plugin at user scope. Re-adding the marketplace moves the pin, so it is
// removed first when present; that uninstalls every plugin from it, so the
// other user-scope plugins from it (such as token-ledger) are installed
// again afterwards. Those that fail, and those at other scopes, which it
// cannot restore, come back as warnings.
func (h *Harness) Install(version string) ([]string, error) {
	src, err := Source(version)
	if err != nil {
		return nil, err
	}
	var warnings, others []string
	if present, err := h.hasMarketplace(); err != nil {
		return nil, err
	} else if present {
		plugins, err := h.plugins()
		if err != nil {
			return nil, err
		}
		for _, p := range plugins {
			if p.ID == pluginID || !strings.HasSuffix(p.ID, "@"+marketplace) {
				continue
			}
			if p.Scope == "user" {
				others = append(others, p.ID)
			} else {
				warnings = append(warnings, fmt.Sprintf("%s (%s scope) was uninstalled with the marketplace; install it again with claude plugin install %s --scope %s", p.ID, p.Scope, p.ID, p.Scope))
			}
		}
		if _, err := h.run("plugin", "marketplace", "remove", marketplace); err != nil {
			return nil, err
		}
	}
	if _, err := h.run("plugin", "marketplace", "add", src); err != nil {
		return nil, err
	}
	out, err := h.run("plugin", "install", pluginID, "--scope", "user", "--json")
	if err != nil {
		return nil, err
	}
	if err := checkResult(out); err != nil {
		return nil, err
	}
	for _, id := range others {
		out, err := h.run("plugin", "install", id, "--scope", "user", "--json")
		if err == nil {
			err = checkResult(out)
		}
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("%s was uninstalled with the marketplace and could not be installed again: %v", id, err))
		}
	}
	return warnings, nil
}

// Remove uninstalls the plugin and removes the marketplace.
func (h *Harness) Remove() error {
	if _, ok, err := h.Installed(); err != nil {
		return err
	} else if ok {
		out, err := h.run("plugin", "uninstall", pluginID, "--scope", "user", "--json")
		if err != nil {
			return err
		}
		if err := checkResult(out); err != nil {
			return err
		}
	}
	if present, err := h.hasMarketplace(); err != nil {
		return err
	} else if present {
		_, err := h.run("plugin", "marketplace", "remove", marketplace)
		return err
	}
	return nil
}

func (h *Harness) hasMarketplace() (bool, error) {
	out, err := h.run("plugin", "marketplace", "list", "--json")
	if err != nil {
		return false, err
	}
	var ms []struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal([]byte(jsonPart(out)), &ms); err != nil {
		return false, fmt.Errorf("reading claude plugin marketplace list: %v", err)
	}
	for _, m := range ms {
		if m.Name == marketplace {
			return true, nil
		}
	}
	return false, nil
}

// checkResult reads the JSON object Claude Code prints as the last line of
// install and uninstall with --json.
func checkResult(out string) error {
	var r struct {
		Outcome string `json:"outcome"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal([]byte(lastLine(out)), &r); err != nil {
		return fmt.Errorf("unexpected output from claude: %s", strings.TrimSpace(out))
	}
	if r.Outcome != "ok" {
		return errors.New(r.Message)
	}
	return nil
}

// jsonPart drops anything Claude Code prints before a JSON document.
func jsonPart(out string) string {
	if k := strings.IndexAny(out, "[{"); k >= 0 {
		return out[k:]
	}
	return lastLine(out)
}

func lastLine(s string) string {
	lines := strings.Split(strings.TrimSpace(s), "\n")
	return lines[len(lines)-1]
}
