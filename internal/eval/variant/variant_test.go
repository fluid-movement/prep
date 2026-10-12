package variant

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func write(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func setup(t *testing.T) Resolver {
	t.Helper()
	root := t.TempDir()
	plugin := filepath.Join(root, "plugin")
	write(t, filepath.Join(plugin, ".claude-plugin", "plugin.json"), `{"name":"prep","version":"1"}`)
	write(t, filepath.Join(plugin, "skills", "prep", "SKILL.md"), "rule 1\nrule 2\n")
	write(t, filepath.Join(plugin, "hooks", "hooks.json"), `{"hooks":{}}`)
	vars := filepath.Join(root, "variants")
	write(t, filepath.Join(vars, "short", "variant.yaml"), "base: current\ndescription: shorter\nreplace: {skills/prep/SKILL.md: SKILL.md}\n")
	write(t, filepath.Join(vars, "short", "SKILL.md"), "rule 1\n")
	write(t, filepath.Join(vars, "nohooks", "variant.yaml"), "base: current\nremove: [hooks/hooks.json]\n")
	write(t, filepath.Join(vars, "stack", "variant.yaml"), "base: short\nadd: {extra.md: extra.md}\nappend_system_prompt: Be brief.\n")
	write(t, filepath.Join(vars, "stack", "extra.md"), "extra")
	write(t, filepath.Join(vars, "prompt-only", "variant.yaml"), "base: none\nappend_system_prompt: Use prep.\n")
	return Resolver{Root: vars, PluginDir: plugin}
}

func TestBuiltins(t *testing.T) {
	r := setup(t)
	cur, err := r.Materialize("current", filepath.Join(t.TempDir(), "p"))
	if err != nil || cur.PluginDir == "" || cur.Hash == "" {
		t.Fatalf("current: %v %+v", err, cur)
	}
	if b, _ := os.ReadFile(filepath.Join(cur.PluginDir, "skills", "prep", "SKILL.md")); string(b) != "rule 1\nrule 2\n" {
		t.Errorf("current skill: %q", b)
	}
	none, err := r.Materialize("none", filepath.Join(t.TempDir(), "p"))
	if err != nil || none.PluginDir != "" {
		t.Fatalf("none: %v %+v", err, none)
	}
	if none.Hash == cur.Hash {
		t.Error("none and current share a hash")
	}
}

func TestHashStableAndSensitive(t *testing.T) {
	r := setup(t)
	mat := func(name string) string {
		v, err := r.Materialize(name, filepath.Join(t.TempDir(), "p"))
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		return v.Hash
	}
	if mat("current") != mat("current") || mat("short") != mat("short") {
		t.Error("hash not stable")
	}
	seen := map[string]string{}
	for _, n := range []string{"current", "none", "short", "nohooks", "stack", "prompt-only"} {
		h := mat(n)
		if prev, dup := seen[h]; dup {
			t.Errorf("%s and %s share hash %s", n, prev, h)
		}
		seen[h] = n
	}
}

func TestOperations(t *testing.T) {
	r := setup(t)
	short, err := r.Materialize("short", filepath.Join(t.TempDir(), "p"))
	if err != nil {
		t.Fatal(err)
	}
	if b, _ := os.ReadFile(filepath.Join(short.PluginDir, "skills", "prep", "SKILL.md")); string(b) != "rule 1\n" {
		t.Errorf("replace: %q", b)
	}
	nh, _ := r.Materialize("nohooks", filepath.Join(t.TempDir(), "p"))
	if _, err := os.Stat(filepath.Join(nh.PluginDir, "hooks", "hooks.json")); !os.IsNotExist(err) {
		t.Error("remove did not remove")
	}
	st, err := r.Materialize("stack", filepath.Join(t.TempDir(), "p"))
	if err != nil {
		t.Fatal(err)
	}
	if b, _ := os.ReadFile(filepath.Join(st.PluginDir, "extra.md")); string(b) != "extra" {
		t.Error("add")
	}
	if b, _ := os.ReadFile(filepath.Join(st.PluginDir, "skills", "prep", "SKILL.md")); string(b) != "rule 1\n" {
		t.Error("stacked variant lost its base's replacement")
	}
	if st.AppendSystemPrompt != "Be brief." {
		t.Errorf("prompt: %q", st.AppendSystemPrompt)
	}
	po, err := r.Materialize("prompt-only", filepath.Join(t.TempDir(), "p"))
	if err != nil || po.PluginDir != "" || po.AppendSystemPrompt != "Use prep." {
		t.Errorf("prompt-only: %v %+v", err, po)
	}
	// the shipped plugin is not modified by any of it
	if b, _ := os.ReadFile(filepath.Join(r.PluginDir, "skills", "prep", "SKILL.md")); string(b) != "rule 1\nrule 2\n" {
		t.Errorf("shipped plugin changed: %q", b)
	}
}

func TestErrors(t *testing.T) {
	r := setup(t)
	bad := func(name, yaml, want string) {
		t.Helper()
		write(t, filepath.Join(r.Root, name, "variant.yaml"), yaml)
		_, err := r.Materialize(name, filepath.Join(t.TempDir(), "p"))
		if err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("%s: error %v, want %q", name, err, want)
		}
	}
	bad("unknown-base", "base: nope\n", "nope")
	bad("no-base", "description: x\n", "no base")
	bad("missing-file", "base: current\nreplace: {nothing/here.md: x.md}\n", "no such file")
	bad("add-existing", "base: current\nadd: {hooks/hooks.json: x.md}\n", "already has it")
	bad("remove-missing", "base: current\nremove: [zzz]\n", "no such file")
	bad("escape", "base: current\nremove: ['../x']\n", "leaves")
	bad("unknown-key", "base: current\nbogus: 1\n", "bogus")
	write(t, filepath.Join(r.Root, "a", "variant.yaml"), "base: b\n")
	write(t, filepath.Join(r.Root, "b", "variant.yaml"), "base: a\n")
	if _, err := r.Materialize("a", filepath.Join(t.TempDir(), "p")); err == nil || !strings.Contains(err.Error(), "cycle") {
		t.Errorf("cycle: %v", err)
	}
}

func TestShippedPlugin(t *testing.T) {
	r := Resolver{Root: t.TempDir(), PluginDir: "../../../plugins/claude-code"}
	v, err := r.Materialize("current", filepath.Join(t.TempDir(), "p"))
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range []string{".claude-plugin/plugin.json", "skills/prep/SKILL.md", "hooks/hooks.json"} {
		if _, err := os.Stat(filepath.Join(v.PluginDir, f)); err != nil {
			t.Errorf("%s: %v", f, err)
		}
	}
}
