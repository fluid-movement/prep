// Package codex integrates prep with Codex by installing its user-scoped
// plugin from the prep plugin marketplace.
package codex

import (
	"encoding/json"
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
	SourceEnv   = "PREP_CODEX_PLUGIN_SOURCE"
)

func init() { setup.Register(&Harness{}) }

// Harness is the Codex integration.
type Harness struct {
	Run func(args ...string) (string, error)
}

func (h *Harness) Name() string  { return "codex" }
func (h *Harness) Title() string { return "Codex" }

func (h *Harness) Detect() bool {
	if h.Run != nil {
		return true
	}
	_, err := exec.LookPath("codex")
	return err == nil
}

func (h *Harness) run(args ...string) (string, error) {
	if h.Run != nil {
		return h.Run(args...)
	}
	out, err := exec.Command("codex", args...).CombinedOutput()
	if err != nil {
		return string(out), fmt.Errorf("codex %s: %v: %s", strings.Join(args, " "), err, strings.TrimSpace(lastLine(string(out))))
	}
	return string(out), nil
}

type installedPlugin struct {
	ID          string `json:"pluginId"`
	Version     string `json:"version"`
	Marketplace string `json:"marketplaceName"`
}

func (h *Harness) plugins() ([]installedPlugin, error) {
	out, err := h.run("plugin", "list", "--json")
	if err != nil {
		return nil, err
	}
	var result struct {
		Installed []installedPlugin `json:"installed"`
	}
	if err := json.Unmarshal([]byte(jsonPart(out)), &result); err != nil {
		return nil, fmt.Errorf("reading codex plugin list: %w", err)
	}
	return result.Installed, nil
}

func (h *Harness) Installed() (string, bool, error) {
	plugins, err := h.plugins()
	if err != nil {
		return "", false, err
	}
	for _, p := range plugins {
		if p.ID == pluginID {
			return "v" + strings.TrimPrefix(p.Version, "v"), true, nil
		}
	}
	return "", false, nil
}

// Source returns a release-pinned GitHub source or a local checkout supplied
// for development installs.
func Source(version string) (string, error) {
	if source := os.Getenv(SourceEnv); source != "" {
		return source, nil
	}
	if !update.IsRelease(version) {
		return "", fmt.Errorf("this is a development build (%s): set %s to a prep checkout, or install a release", version, SourceEnv)
	}
	return "fluid-movement/prep", nil
}

func (h *Harness) marketplacePresent() (bool, error) {
	out, err := h.run("plugin", "marketplace", "list", "--json")
	if err != nil {
		return false, err
	}
	var result struct {
		Marketplaces []struct {
			Name string `json:"name"`
		} `json:"marketplaces"`
	}
	if err := json.Unmarshal([]byte(jsonPart(out)), &result); err != nil {
		return false, fmt.Errorf("reading codex marketplace list: %w", err)
	}
	for _, m := range result.Marketplaces {
		if m.Name == marketplace {
			return true, nil
		}
	}
	return false, nil
}

// Install adds or refreshes the release-pinned marketplace and plugin. Other
// installed plugins from the prep marketplace are restored after its pin is
// moved, with warnings if Codex can no longer install one of them.
func (h *Harness) Install(version string) ([]string, error) {
	source, err := Source(version)
	if err != nil {
		return nil, err
	}
	if current, ok, err := h.Installed(); err != nil {
		return nil, err
	} else if ok && current == version {
		return nil, nil
	}

	plugins, err := h.plugins()
	if err != nil {
		return nil, err
	}
	var others []string
	for _, p := range plugins {
		if p.Marketplace == marketplace && p.ID != pluginID {
			others = append(others, p.ID)
		}
	}
	present, err := h.marketplacePresent()
	if err != nil {
		return nil, err
	}
	if present {
		if _, err := h.run("plugin", "marketplace", "remove", marketplace, "--json"); err != nil {
			return nil, err
		}
	}
	add := []string{"plugin", "marketplace", "add", source}
	if os.Getenv(SourceEnv) == "" {
		add = append(add, "--ref", version, "--sparse", ".agents/plugins", "--sparse", "plugins/codex")
	}
	add = append(add, "--json")
	if _, err := h.run(add...); err != nil {
		return nil, err
	}
	if _, err := h.run("plugin", "add", pluginID, "--json"); err != nil {
		return nil, err
	}
	var warnings []string
	for _, id := range others {
		if _, err := h.run("plugin", "add", id, "--json"); err != nil {
			warnings = append(warnings, fmt.Sprintf("%s was removed with the prep marketplace and could not be installed again: %v", id, err))
		}
	}
	warnings = append(warnings, "Codex requires reviewing and trusting the prep plugin hooks before they run; open /hooks in Codex")
	return warnings, nil
}

func (h *Harness) Remove() error {
	plugins, err := h.plugins()
	if err != nil {
		return err
	}
	installed := false
	others := false
	for _, p := range plugins {
		if p.ID == pluginID {
			installed = true
		} else if p.Marketplace == marketplace {
			others = true
		}
	}
	if installed {
		if _, err := h.run("plugin", "remove", pluginID, "--json"); err != nil {
			return err
		}
	}
	if others {
		return nil
	}
	present, err := h.marketplacePresent()
	if err != nil || !present {
		return err
	}
	_, err = h.run("plugin", "marketplace", "remove", marketplace, "--json")
	return err
}

func jsonPart(s string) string {
	if i := strings.IndexAny(s, "[{\""); i > 0 {
		return s[i:]
	}
	return s
}

func lastLine(s string) string {
	lines := strings.Split(strings.TrimSpace(s), "\n")
	if len(lines) == 0 {
		return ""
	}
	return lines[len(lines)-1]
}
