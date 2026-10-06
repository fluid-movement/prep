package claudecode

import (
	"encoding/json"
	"strings"
	"testing"
)

// fakeClaude mimics the claude plugin commands prep uses, with state.
type fakeClaude struct {
	marketplace string // source, empty when not added
	installed   string // version, empty when not installed
	calls       []string
	failInstall bool
}

func (f *fakeClaude) run(args ...string) (string, error) {
	cmd := strings.Join(args, " ")
	f.calls = append(f.calls, cmd)
	switch {
	case cmd == "plugin list --json":
		ps := []map[string]any{}
		if f.installed != "" {
			ps = append(ps, map[string]any{"id": "prep@prep", "version": f.installed, "scope": "user"})
		}
		b, _ := json.Marshal(ps)
		return "warming up\n" + string(b), nil
	case cmd == "plugin marketplace list --json":
		ms := []map[string]any{}
		if f.marketplace != "" {
			ms = append(ms, map[string]any{"name": "prep"})
		}
		b, _ := json.Marshal(ms)
		return string(b), nil
	case strings.HasPrefix(cmd, "plugin marketplace add "):
		f.marketplace = strings.TrimPrefix(cmd, "plugin marketplace add ")
		return "✔ Successfully added marketplace: prep", nil
	case cmd == "plugin marketplace remove prep":
		f.marketplace, f.installed = "", ""
		return "removed", nil
	case cmd == "plugin install prep@prep --scope user --json":
		if f.failInstall {
			return `{"outcome":"failed","message":"Plugin not found"}`, nil
		}
		f.installed = strings.TrimPrefix(f.marketplace[strings.LastIndex(f.marketplace, "#")+1:], "v")
		return `{"outcome":"ok","message":"Successfully installed plugin: prep@prep (scope: user)"}`, nil
	case cmd == "plugin uninstall prep@prep --scope user --json":
		f.installed = ""
		return `{"outcome":"ok"}`, nil
	}
	return "", nil
}

func TestInstallUpdateRemove(t *testing.T) {
	t.Setenv(SourceEnv, "")
	f := &fakeClaude{}
	h := &Harness{Run: f.run}

	if v, ok, err := h.Installed(); err != nil || ok {
		t.Fatalf("installed before install: %s %v %v", v, ok, err)
	}
	if err := h.Install("v0.2.0"); err != nil {
		t.Fatal(err)
	}
	if f.marketplace != "fluid-movement/prep#v0.2.0" {
		t.Fatalf("marketplace pinned to %q", f.marketplace)
	}
	if v, ok, _ := h.Installed(); !ok || v != "v0.2.0" {
		t.Fatalf("installed = %s %v", v, ok)
	}

	// Updating re-pins the marketplace.
	f.calls = nil
	if err := h.Install("v0.3.0"); err != nil {
		t.Fatal(err)
	}
	if f.calls[1] != "plugin marketplace remove prep" || f.marketplace != "fluid-movement/prep#v0.3.0" {
		t.Fatalf("update calls %v, pin %q", f.calls, f.marketplace)
	}

	if err := h.Remove(); err != nil {
		t.Fatal(err)
	}
	if f.installed != "" || f.marketplace != "" {
		t.Fatalf("after remove: %+v", f)
	}
	if err := h.Remove(); err != nil {
		t.Fatalf("removing twice: %v", err)
	}
}

func TestSourceAndFailures(t *testing.T) {
	t.Setenv(SourceEnv, "")
	if _, err := Source("v0.0.0-20261006090120-93aa83c32597"); err == nil || !strings.Contains(err.Error(), SourceEnv) {
		t.Fatalf("development build: %v", err)
	}
	t.Setenv(SourceEnv, "/src/prep")
	if s, _ := Source("dev"); s != "/src/prep" {
		t.Fatalf("override: %q", s)
	}
	f := &fakeClaude{failInstall: true}
	if err := (&Harness{Run: f.run}).Install("v1.0.0"); err == nil || err.Error() != "Plugin not found" {
		t.Fatalf("failed install: %v", err)
	}
}
