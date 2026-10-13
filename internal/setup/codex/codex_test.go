package codex

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

type fakeCodex struct {
	plugins     map[string]installedPlugin
	marketplace bool
	version     string
	calls       []string
}

func (f *fakeCodex) run(args ...string) (string, error) {
	command := strings.Join(args, " ")
	f.calls = append(f.calls, command)
	switch {
	case command == "plugin list --json":
		list := []installedPlugin{}
		for _, p := range f.plugins {
			list = append(list, p)
		}
		b, _ := json.Marshal(map[string]any{"installed": list})
		return string(b), nil
	case command == "plugin marketplace list --json":
		list := []map[string]string{}
		if f.marketplace {
			list = append(list, map[string]string{"name": "prep"})
		}
		b, _ := json.Marshal(map[string]any{"marketplaces": list})
		return string(b), nil
	case command == "plugin marketplace remove prep --json":
		f.marketplace = false
		for id, p := range f.plugins {
			if p.Marketplace == marketplace {
				delete(f.plugins, id)
			}
		}
	case strings.HasPrefix(command, "plugin marketplace add "):
		f.marketplace = true
		for i, arg := range args {
			if arg == "--ref" {
				f.version = strings.TrimPrefix(args[i+1], "v")
			}
		}
	case strings.HasPrefix(command, "plugin add "):
		id := args[2]
		f.plugins[id] = installedPlugin{ID: id, Marketplace: marketplace, Version: f.version}
	case strings.HasPrefix(command, "plugin remove "):
		delete(f.plugins, args[2])
	default:
		return "", fmt.Errorf("unexpected command %s", command)
	}
	return "{}", nil
}

func TestCodexRefreshAndRemovePreserveOtherPlugins(t *testing.T) {
	t.Setenv(SourceEnv, "")
	sibling := installedPlugin{ID: "sibling@prep", Marketplace: "prep", Version: "0.2.0"}
	remote := installedPlugin{ID: "other@remote", Marketplace: "remote", Version: "1.2.3"}
	f := &fakeCodex{marketplace: true, version: "0.2.0", plugins: map[string]installedPlugin{
		pluginID: {ID: pluginID, Marketplace: "prep", Version: "0.2.0"}, sibling.ID: sibling, remote.ID: remote,
	}}
	h := &Harness{Run: f.run}
	if warnings, err := h.Install("v0.3.0"); err != nil || len(warnings) != 1 {
		t.Fatalf("install: %v %v", warnings, err)
	}
	if v, ok, err := h.Installed(); err != nil || !ok || v != "v0.3.0" {
		t.Fatalf("installed: %s %v %v", v, ok, err)
	}
	if f.plugins[sibling.ID].ID != sibling.ID || f.plugins[remote.ID] != remote {
		t.Fatal("refresh disturbed another plugin")
	}
	if !strings.Contains(strings.Join(f.calls, "\n"), "--ref v0.3.0 --sparse .agents/plugins --sparse plugins/codex") {
		t.Fatal("release source not pinned")
	}
	if err := h.Remove(); err != nil {
		t.Fatal(err)
	}
	if _, ok := f.plugins[pluginID]; ok || !f.marketplace || f.plugins[remote.ID] != remote || f.plugins[sibling.ID].ID != sibling.ID {
		t.Fatal("remove disturbed another plugin or marketplace")
	}
	delete(f.plugins, sibling.ID)
	if err := h.Remove(); err != nil || f.marketplace || f.plugins[remote.ID] != remote {
		t.Fatalf("unused marketplace cleanup: %v", err)
	}
}

func TestCodexLocalDevelopmentSourceAndNoopRefresh(t *testing.T) {
	t.Setenv(SourceEnv, "/local/checkout")
	f := &fakeCodex{plugins: map[string]installedPlugin{}, version: "0.3.0"}
	h := &Harness{Run: f.run}
	if _, err := h.Install("v0.3.0"); err != nil {
		t.Fatal(err)
	}
	joined := strings.Join(f.calls, "\n")
	if !strings.Contains(joined, "plugin marketplace add /local/checkout --json") || strings.Contains(joined, "--ref") {
		t.Fatalf("development source: %s", joined)
	}
	f.calls = nil
	if warnings, err := h.Install("v0.3.0"); err != nil || len(warnings) != 0 || len(f.calls) != 1 {
		t.Fatalf("noop: %v %v %v", warnings, err, f.calls)
	}
}
